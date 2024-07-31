package fts

import (
	"context"
	"database/sql"
	"github.com/knaka/biblioseeq/db/sqlcgen"
	ftslog "github.com/knaka/biblioseeq/log"
	. "github.com/knaka/go-utils"
	"github.com/rjeczalik/notify"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type DirIndexer struct {
	path           string
	fileExtensions []string
	indexMutex     sync.Mutex
	dbConn         *sql.DB
	chStop         chan any
	chStarted      chan any
	chScanned      chan any
}

func NewDirIndexer(dbConn *sql.DB, path string, fileExtensions []string) *DirIndexer {
	return &DirIndexer{
		dbConn:         dbConn,
		chStop:         make(chan any),
		chStarted:      make(chan any, 2),
		chScanned:      make(chan any, 2),
		path:           path,
		fileExtensions: fileExtensions,
	}
}

func (dirIndexer *DirIndexer) IndexFile(path string) {
	ctx := context.Background()
	store := sqlcgen.New(dirIndexer.dbConn)
	if lo.NoneBy(dirIndexer.fileExtensions, func(ext string) bool {
		return filepath.Ext(path) == ext
	}) {
		return
	}
	stat, err := os.Stat(path)
	// broken symlink?
	if err != nil {
		return
	}
	if stat.IsDir() {
		return
	}
	body := tokenizeJapanese(string(V(os.ReadFile(path))))
	if file := R(store.GetFile(ctx, &sqlcgen.GetFileParams{Path: path})).NilIf(sql.ErrNoRows); file == nil {
		ftslog.Println("Adding index for file:", path)
		idFtsFile := V(store.AddFtsFile(ctx, &sqlcgen.AddFtsFileParams{
			Body: body,
		}))
		V0(store.AddFile(context.Background(), &sqlcgen.AddFileParams{
			Path:       path,
			FtsFileID:  idFtsFile,
			ModifiedAt: stat.ModTime(),
			Size:       stat.Size(),
		}))
	} else {
		if file.ModifiedAt.Equal(stat.ModTime()) && file.Size == stat.Size() {
			ftslog.Println("Skipped file:", path)
			return
		}
		ftslog.Println("Updating index for file:", path)
		V0(store.UpdateFtsFile(context.Background(), &sqlcgen.UpdateFtsFileParams{
			Path: path,
			Body: body,
		}))
		V0(store.UpdateFile(context.Background(), &sqlcgen.UpdateFileParams{
			Path:       path,
			ModifiedAt: stat.ModTime(),
			Size:       stat.Size(),
		}))
	}
}

func (dirIndexer *DirIndexer) DeleteFileOrDirectoryRecursively(path string) {
	dirIndexer.indexMutex.Lock()
	defer dirIndexer.indexMutex.Unlock()
	store := sqlcgen.New(dirIndexer.dbConn)
	ctx := context.Background()
	V0(store.DeleteFtsFiles(ctx, &sqlcgen.DeleteFtsFilesParams{Path: path}))
	V0(store.DeleteFiles(ctx, &sqlcgen.DeleteFilesParams{Path: path}))
}

func (dirIndexer *DirIndexer) onEvent(eventInfo notify.EventInfo) {
	dirIndexer.indexMutex.Lock()
	defer dirIndexer.indexMutex.Unlock()
	event := eventInfo.Event()
	path := eventInfo.Path()
	//sys := eventInfo.Sys()
	if event&notify.Write != 0 {
		// ファイルへの書き込み。単体更新
		dirIndexer.IndexFile(path)
	}
	if event&notify.Remove != 0 {
		// log.Println("ファイルもしくはディレクトリの削除。パス前方一致削除。再帰で通知が来ていたら、配下のファイルはすでに消されていると思われるので、何もしなくて良いこともある")
		// 消えたのがディレクトリかは分からない。もう無いので
		dirIndexer.DeleteFileOrDirectoryRecursively(path)
	}
	if event&notify.Rename != 0 {
		stat := R(os.Stat(path)).NilIf(os.ErrNotExist)
		if stat == nil {
			dirIndexer.DeleteFileOrDirectoryRecursively(path)
		}
		if stat.IsDir() {
			dirIndexer.indexDirectory(path)
		} else {
			dirIndexer.IndexFile(path)
		}
	}
}

func (dirIndexer *DirIndexer) indexFile(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	dirIndexer.IndexFile(path)
	return nil
}

func (dirIndexer *DirIndexer) removeNotExistingFiles() {
	ctx := context.Background()
	store := sqlcgen.New(dirIndexer.dbConn)
	files, err := store.GetFiles(ctx)
	if err != nil {
		return
	}
	for _, file := range files {
		var shouldRemove bool
		if lo.NoneBy(dirIndexer.fileExtensions, func(ext string) bool {
			return filepath.Ext(file.Path) == ext
		}) {
			shouldRemove = true
		}
		if _, err := os.Stat(file.Path); err != nil {
			shouldRemove = true
		}
		if shouldRemove {
			ftslog.Println("Removing index for file:", file.Path)
			path := file.Path
			V0(store.DeleteFtsFiles(ctx, &sqlcgen.DeleteFtsFilesParams{Path: path}))
			V0(store.DeleteFiles(ctx, &sqlcgen.DeleteFilesParams{Path: path}))
		}
	}
}

func (dirIndexer *DirIndexer) indexDirectory(path string) {
	V0(filepath.Walk(path, dirIndexer.indexFile))
}

func (dirIndexer *DirIndexer) IndexAll() {
	dirIndexer.indexMutex.Lock()
	defer dirIndexer.indexMutex.Unlock()
	dirIndexer.removeNotExistingFiles()
	dirIndexer.indexDirectory(dirIndexer.path)
	dirIndexer.chScanned <- "scanned"
}

func (dirIndexer *DirIndexer) WatchContinuously(ctx context.Context) {
	ch := make(chan notify.EventInfo, 10)
	V0(notify.Watch(filepath.Join(dirIndexer.path, "..."), ch,
		notify.Write, notify.Remove, notify.Rename))
	defer notify.Stop(ch)
	defer close(ch)
	dirIndexer.chStarted <- "started"
outer:
	for {
		select {
		case eventInfo := <-ch:
			dirIndexer.onEvent(eventInfo)
		case <-dirIndexer.chStop:
			break outer
		case <-ctx.Done():
			break outer
		}
	}
}

func (dirIndexer *DirIndexer) WaitForWatchingStarted(ctx context.Context) {
	select {
	case started := <-dirIndexer.chStarted:
		dirIndexer.chStarted <- started
	case <-ctx.Done():
	}

}

func (dirIndexer *DirIndexer) StopWatching() {
	dirIndexer.chStop <- "stop"
}

func (dirIndexer *DirIndexer) Watching() bool {
	return len(dirIndexer.chStop) == 0 && len(dirIndexer.chStarted) > 0
}

func (dirIndexer *DirIndexer) WaitForInitialScanFinished(ctx context.Context) {
	select {
	case scanned := <-dirIndexer.chScanned:
		dirIndexer.chScanned <- scanned
	case <-ctx.Done():
	}
}

func (dirIndexer *DirIndexer) InitialScanFinished() bool {
	return len(dirIndexer.chScanned) > 0
}
