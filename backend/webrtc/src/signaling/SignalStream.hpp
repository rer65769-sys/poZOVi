#pragma once

#include <atomic>
#include <memory>
#include <functional>
#include "../dispatcher/Dispatcher.hpp"
#include "../session/SessionManager.hpp"
#include "ISignalingSink.hpp"
#include "webrtc_service.grpc.pb.h"

namespace webrtc::signaling {

// SignalStream functions must be protected and therefore called via Dispatcher
class SignalStream final : ISignalingSink, std::enable_shared_from_this<SignalStream> {
public:
    SignalStream(dispatch::Dispatcher& dispatcher, session::SessionManager& sessionManager,
        grpc::ServerReaderWriter<webrtc::SignalingMessage, webrtc::SignalingMessage>* stream)
        : dispatcher_(dispatcher), sessionManager_(sessionManager), stream_(stream) {};
    
    void start();
    void send(const webrtc::SignalingMessage& message);
    void close();

    void sendLocalDescription(const Description& desc, const uint32_t id) override;
    void sendIceCandidate(const IceCandidate& candidate, const uint32_t id) override;
    void sendConnectionState(const ConnectionState state, const uint32_t id) override;
private:
    dispatch::Dispatcher& dispatcher_;
    session::SessionManager& sessionManager_;
    grpc::ServerReaderWriter<webrtc::SignalingMessage, webrtc::SignalingMessage>* stream_;
    std::atomic<bool> closed_{false};

    void handleRemoteDescription(const webrtc::SignalingMessage& message);
    void handleIceCandidate(const webrtc::SignalingMessage& message);
    void handleStateChange(const webrtc::SignalingMessage& message);
    template<typename Fn> void withSession(uint32_t sessionId, Fn&& fn);
};

}