// @generated by protoc-gen-es v1.10.0 with parameter "target=js+dts"
// @generated from file v1/error.proto (package v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

/**
 * @generated from enum v1.ErrorCode
 */
export declare enum ErrorCode {
  /**
   * Zero value is required and UNSPECIFIED is a convention.
   * Enum value names should be unique since they are global.
   *
   * ゼロは必須で UNSPECIFIED にする慣習。デフォルト値なので
   * enum 値名は unique である必要がある。プレフィクスをつけた方が良くはある
   *
   * @generated from enum value: ERROR_CODE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * よく分からない時はこれを返せ
   *
   * @generated from enum value: ERROR_CODE_INTERNAL = 1;
   */
  INTERNAL = 1,

  /**
   * ログイン
   *
   * @generated from enum value: ERROR_CODE_USER_NOT_EXISTS = 20000;
   */
  USER_NOT_EXISTS = 20000,

  /**
   * 認証
   *
   * @generated from enum value: ERROR_CODE_TOKEN_INVALID = 20102;
   */
  TOKEN_INVALID = 20102,

  /**
   * @generated from enum value: ERROR_CODE_TOKEN_EXPIRED = 20101;
   */
  TOKEN_EXPIRED = 20101,

  /**
   * 認可
   *
   * @generated from enum value: ERROR_CODE_PERMISSION_DENIED = 20201;
   */
  PERMISSION_DENIED = 20201,

  /**
   * その他
   *
   * @generated from enum value: ERROR_CODE_INVALID_ARGUMENT = 20301;
   */
  INVALID_ARGUMENT = 20301,
}

