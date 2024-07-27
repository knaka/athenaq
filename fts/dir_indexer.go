package fts

import (
	"context"
	"database/sql"
	"github.com/knaka/biblioseeq/db/sqlcgen"
	. "github.com/knaka/go-utils"
	"github.com/rjeczalik/notify"
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
	eventCh        chan notify.EventInfo
	watcherCh      chan any
	watchingCh     chan any
	scannedCh      chan any
}

func NewDirIndexer(dbConn *sql.DB, path string, fileExtensions []string) *DirIndexer {
	return &DirIndexer{
		dbConn:         dbConn,
		eventCh:        make(chan notify.EventInfo),
		watcherCh:      make(chan any),
		watchingCh:     make(chan any, 2),
		scannedCh:      make(chan any, 2),
		path:           path,
		fileExtensions: fileExtensions,
	}
}

func (dirIndexer *DirIndexer) IndexFile(path string) {
	ctx := context.Background()
	store := sqlcgen.New(dirIndexer.dbConn)
	stat := V(os.Stat(path))
	body := tokenizeJapanese(string(V(os.ReadFile(path))))
	if file := PR(store.GetFile(ctx, &sqlcgen.GetFileParams{Path: path})).NilIf(sql.ErrNoRows); file == nil {
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

func (dirIndexer *DirIndexer) indexDirectory(path string) {
	V0(filepath.Walk(path, dirIndexer.indexFile))
}

func (dirIndexer *DirIndexer) IndexAll() {
	dirIndexer.indexMutex.Lock()
	defer dirIndexer.indexMutex.Unlock()
	dirIndexer.indexDirectory(dirIndexer.path)
	dirIndexer.scannedCh <- "scanned"
}

func (dirIndexer *DirIndexer) WatchContinuously() {
	ch := make(chan notify.EventInfo)
	V0(notify.Watch(filepath.Join(dirIndexer.path, "..."), ch,
		notify.Write, notify.Remove, notify.Rename))
	defer notify.Stop(ch)
	defer close(ch)
	dirIndexer.watchingCh <- "watching"
root:
	for {
		select {
		case eventInfo := <-ch:
			dirIndexer.onEvent(eventInfo)
		case <-dirIndexer.watcherCh:
			break root
		}
	}
}

func (dirIndexer *DirIndexer) WaitForWatchingStarted() {
	<-dirIndexer.watchingCh
	dirIndexer.watchingCh <- "watching"
}

func (dirIndexer *DirIndexer) StopWatching() {
	close(dirIndexer.watcherCh)
}

func (dirIndexer *DirIndexer) WaitForInitialScanFinished() {
	<-dirIndexer.scannedCh
	dirIndexer.scannedCh <- "scanned"
}