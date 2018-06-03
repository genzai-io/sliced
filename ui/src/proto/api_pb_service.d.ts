// package: api
// file: proto/api.proto

import * as proto_api_pb from "../proto/api_pb";
import {grpc} from "grpc-web-client";

type APIServiceAuth = {
  readonly methodName: string;
  readonly service: typeof APIService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof proto_api_pb.AuthRequest;
  readonly responseType: typeof proto_api_pb.AuthReply;
};

type APIServiceRegister = {
  readonly methodName: string;
  readonly service: typeof APIService;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof proto_api_pb.RegisterRequest;
  readonly responseType: typeof proto_api_pb.RegisterReply;
};

type APIServiceEvents = {
  readonly methodName: string;
  readonly service: typeof APIService;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof proto_api_pb.EventsRequest;
  readonly responseType: typeof proto_api_pb.Event;
};

export class APIService {
  static readonly serviceName: string;
  static readonly Auth: APIServiceAuth;
  static readonly Register: APIServiceRegister;
  static readonly Events: APIServiceEvents;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }
export type ServiceClientOptions = { transport: grpc.TransportConstructor }

interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: () => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}

export class APIServiceClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: ServiceClientOptions);
  auth(
    requestMessage: proto_api_pb.AuthRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: proto_api_pb.AuthReply|null) => void
  ): void;
  auth(
    requestMessage: proto_api_pb.AuthRequest,
    callback: (error: ServiceError, responseMessage: proto_api_pb.AuthReply|null) => void
  ): void;
  register(
    requestMessage: proto_api_pb.RegisterRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError, responseMessage: proto_api_pb.RegisterReply|null) => void
  ): void;
  register(
    requestMessage: proto_api_pb.RegisterRequest,
    callback: (error: ServiceError, responseMessage: proto_api_pb.RegisterReply|null) => void
  ): void;
  events(requestMessage: proto_api_pb.EventsRequest, metadata?: grpc.Metadata): ResponseStream<proto_api_pb.Event>;
}

