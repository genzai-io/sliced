// package: api
// file: proto/api.proto

var proto_api_pb = require("../proto/api_pb");
var grpc = require("grpc-web-client").grpc;

var APIService = (function () {
  function APIService() {}
  APIService.serviceName = "api.APIService";
  return APIService;
}());

APIService.Auth = {
  methodName: "Auth",
  service: APIService,
  requestStream: false,
  responseStream: false,
  requestType: proto_api_pb.AuthRequest,
  responseType: proto_api_pb.AuthReply
};

APIService.Register = {
  methodName: "Register",
  service: APIService,
  requestStream: false,
  responseStream: false,
  requestType: proto_api_pb.RegisterRequest,
  responseType: proto_api_pb.RegisterReply
};

APIService.Events = {
  methodName: "Events",
  service: APIService,
  requestStream: false,
  responseStream: true,
  requestType: proto_api_pb.EventsRequest,
  responseType: proto_api_pb.Event
};

exports.APIService = APIService;

function APIServiceClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

APIServiceClient.prototype.auth = function auth(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(APIService.Auth, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

APIServiceClient.prototype.register = function register(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  grpc.unary(APIService.Register, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          callback(Object.assign(new Error(response.statusMessage), { code: response.status, metadata: response.trailers }), null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
};

APIServiceClient.prototype.events = function events(requestMessage, metadata) {
  var listeners = {
    data: [],
    end: [],
    status: []
  };
  var client = grpc.invoke(APIService.Events, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    onMessage: function (responseMessage) {
      listeners.data.forEach(function (handler) {
        handler(responseMessage);
      });
    },
    onEnd: function (status, statusMessage, trailers) {
      listeners.end.forEach(function (handler) {
        handler();
      });
      listeners.status.forEach(function (handler) {
        handler({ code: status, details: statusMessage, metadata: trailers });
      });
      listeners = null;
    }
  });
  return {
    on: function (type, handler) {
      listeners[type].push(handler);
      return this;
    },
    cancel: function () {
      listeners = null;
      client.close();
    }
  };
};

exports.APIServiceClient = APIServiceClient;

