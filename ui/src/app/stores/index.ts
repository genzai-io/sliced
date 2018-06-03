import {grpc} from "grpc-web-client";
import UnaryOutput = grpc.UnaryOutput;

export * from './RouterStore'
export * from './createStore'
export * from './APIStore'
export * from './LayoutStore'

export function isOk(reply: UnaryOutput<any>): boolean {
  return reply.status === grpc.Code.OK && reply.message
}
