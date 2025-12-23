#pragma once

#include "webrtc_service.grpc.pb.h"
#include <grpc++/grpc++.h>
#include <memory>

#include "../session/SessionManager.hpp"
#include "../dispatcher/Dispatcher.hpp"

namespace webrtc::rpc {

class WebRTCServiceImpl final : public webrtc::WebRTCService::Service {
public:
    WebRTCServiceImpl(session::SessionManager& sessionManager, dispatch::Dispatcher& dispatcher) 
        : sessionManager_(sessionManager), dispatcher_(dispatcher) {}

    grpc::Status Signal(
        grpc::ServerContext* context,
        grpc::ServerReaderWriter<webrtc::SignalingMessage, webrtc::SignalingMessage>* stream
    ) override;
    
private:
    session::SessionManager& sessionManager_;
    dispatch::Dispatcher& dispatcher_;
};

}