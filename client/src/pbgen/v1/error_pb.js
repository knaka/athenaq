// @generated by protoc-gen-es v1.10.0 with parameter "target=js+dts"
// @generated from file v1/error.proto (package v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum v1.ErrorCode
 */
export const ErrorCode = /*@__PURE__*/ proto3.makeEnum(
  "v1.ErrorCode",
  [
    {no: 0, name: "ERROR_CODE_UNSPECIFIED", localName: "UNSPECIFIED"},
    {no: 1, name: "ERROR_CODE_INTERNAL", localName: "INTERNAL"},
    {no: 20000, name: "ERROR_CODE_USER_NOT_EXISTS", localName: "USER_NOT_EXISTS"},
    {no: 20102, name: "ERROR_CODE_TOKEN_INVALID", localName: "TOKEN_INVALID"},
    {no: 20101, name: "ERROR_CODE_TOKEN_EXPIRED", localName: "TOKEN_EXPIRED"},
    {no: 20201, name: "ERROR_CODE_PERMISSION_DENIED", localName: "PERMISSION_DENIED"},
    {no: 20301, name: "ERROR_CODE_INVALID_ARGUMENT", localName: "INVALID_ARGUMENT"},
  ],
);

