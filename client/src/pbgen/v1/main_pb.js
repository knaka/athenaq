// Language Guide (proto 3) | Protocol Buffers Documentation https://protobuf.dev/programming-guides/proto3/

// @generated by protoc-gen-es v1.10.0 with parameter "target=js+dts"
// @generated from file v1/main.proto (package v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from message v1.VersionInfo
 */
export const VersionInfo = /*@__PURE__*/ proto3.makeMessageType(
  "v1.VersionInfo",
  () => [
    { no: 1, name: "version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message v1.GetVersionInfoRequest
 */
export const GetVersionInfoRequest = /*@__PURE__*/ proto3.makeMessageType(
  "v1.GetVersionInfoRequest",
  [],
);

/**
 * @generated from message v1.GetVersionInfoResponse
 */
export const GetVersionInfoResponse = /*@__PURE__*/ proto3.makeMessageType(
  "v1.GetVersionInfoResponse",
  () => [
    { no: 1, name: "version_info", kind: "message", T: VersionInfo },
  ],
);
