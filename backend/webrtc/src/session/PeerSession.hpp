#pragma once
#include <mutex>
#include <memory>
#include <expected>
#include <functional>
#include "../rtc/RtcPeer.hpp"
#include "../signaling/ISignalingSink.hpp"
#include "../signaling/SignalingTypes.hpp"
#include "../dispatcher/Dispatcher.hpp"

namespace webrtc::session {

enum class AttachSignalStreamError {
    SessionClosed,
    InvalidSignalStream
};

enum class SetRemoteDescriptionError {
    SessionClosed,
    InvalidDescription,
    PeerConnectionNotStarted
};

enum class IceCandidateError {
    SessionClosed,
    InvalidCandidate,
    RemoteDescriptionNotSet
};

class PeerSession : public std::enable_shared_from_this<PeerSession> {
public:
    PeerSession(uint32_t id, dispatch::Dispatcher& dispatcher, std::function<void()> callback) 
        : sessionId_(id), dispatcher_(dispatcher), onSessionTerminated(callback) {};
    ~PeerSession() { close(); };
    void start();
    void close();
    auto attachSignalStream(const std::shared_ptr<signaling::ISignalingSink>& signalStream) -> std::expected<void, AttachSignalStreamError>;
    auto setRemoteDescription(const signaling::Description& desc) -> std::expected<void, SetRemoteDescriptionError>;
    auto addIceCandidate(const signaling::IceCandidate& candidate) -> std::expected<void, IceCandidateError>;
    void handleConnectionStateChange(signaling::ConnectionState newState);
    
private:
    std::mutex mutex_;
    bool isClosed_ = false;
    std::shared_ptr<rtc::RtcPeer> rtcPeer_;
    std::shared_ptr<signaling::ISignalingSink> signalStream_;
    uint32_t sessionId_;
    dispatch::Dispatcher& dispatcher_;
    signaling::ConnectionState state_;

    std::function<void()> onSessionTerminated;
    void setCallbacks();
    void handleLocalDescription(const signaling::Description& localDesc);
    void handleIceCandidate(const signaling::IceCandidate& candidate);
    bool isValidStateTransition(signaling::ConnectionState from, signaling::ConnectionState to);
    template <typename Fn> void postToDispatcher(Fn&& fn);
};

}