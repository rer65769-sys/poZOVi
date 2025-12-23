#include <iostream>
#include <memory>
#include <string>

#include <grpcpp/grpcpp.h>

#include "dispatcher/Dispatcher.hpp"
#include "session/SessionManager.hpp"
#include "service/WebRTCServiceImpl.hpp"

#include "webrtc_service.grpc.pb.h"

int main(int argc, char** argv) {
    const std::string serverAddress = "0.0.0.0:50051";

    webrtc::dispatch::Dispatcher dispatcher;
    dispatcher.start();

    webrtc::session::SessionManager sessionManager(dispatcher);

    webrtc::rpc::WebRTCServiceImpl service(sessionManager, dispatcher);

    // gRPC server setup
    grpc::ServerBuilder builder;
    builder.AddListeningPort(serverAddress, grpc::InsecureServerCredentials());
    builder.RegisterService(&service);

    std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
    if (!server) {
        std::cerr << "Failed to start gRPC server\n";
        return 1;
    }

    std::cout << "WebRTC signaling server listening on " << serverAddress << std::endl;

    // Block until shutdown
    server->Wait();

    return 0;
}