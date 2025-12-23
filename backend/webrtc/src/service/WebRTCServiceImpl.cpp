#include "WebRTCServiceImpl.hpp"
#include "../signaling/SignalStream.hpp"

using namespace webrtc::rpc;

grpc::Status WebRTCServiceImpl::Signal(grpc::ServerContext* context, 
    grpc::ServerReaderWriter<webrtc::SignalingMessage, webrtc::SignalingMessage>* stream) {
    
    auto signalingStream = std::make_shared<signaling::SignalStream>(
        dispatcher_,
        sessionManager_,
        stream
    );

    signalingStream->start();
    signalingStream->close();

    return grpc::Status::OK;
}