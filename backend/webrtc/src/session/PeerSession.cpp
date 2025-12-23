#include "PeerSession.hpp"

using namespace webrtc::session;

void PeerSession::start() {
    std::lock_guard<std::mutex> lock(mutex_);

    if (isClosed_) return;
    if (rtcPeer_) return;
    rtcPeer_ = std::make_shared<rtc::RtcPeer>();

    setCallbacks();

    rtcPeer_->start();
    state_ = signaling::ConnectionState::New;
}

void PeerSession::setCallbacks() {
    auto weakSelf = weak_from_this();

    rtcPeer_->onLocalDescription = [weakSelf](const signaling::Description& localDesc) {
        if (auto self = weakSelf.lock()) {
            self->handleLocalDescription(localDesc);
        }
    };
    
    rtcPeer_->onIceCandidate = [weakSelf](const signaling::IceCandidate& candidate) {
        if (auto self = weakSelf.lock()) {
            self->handleIceCandidate(candidate);
        }
    };

    rtcPeer_->onConnectionStateChange = [weakSelf](signaling::ConnectionState state) {
        if (auto self = weakSelf.lock()) {
            self->handleConnectionStateChange(state);
        }
    };
}

template <typename Fn>
void PeerSession::postToDispatcher(Fn&& fn) {
    auto weakSelf = weak_from_this();
    dispatcher_.post([weakSelf, fn = std::move(fn)]() mutable {
        if (auto self = weakSelf.lock()) {
            fn(*self);
        }
    });
}

void PeerSession::handleLocalDescription(const signaling::Description& localDesc) {
    if (isClosed_) return;
    if (signalStream_) {
        postToDispatcher([localDesc](PeerSession& self){
            self.signalStream_->sendLocalDescription(localDesc, self.sessionId_);
        });
    }
}

void PeerSession::handleIceCandidate(const signaling::IceCandidate& candidate) {
    if (isClosed_) return;
    if (signalStream_) {
        postToDispatcher([candidate](PeerSession& self){
            self.signalStream_->sendIceCandidate(candidate, self.sessionId_);
        });
    }
}

void PeerSession::handleConnectionStateChange(signaling::ConnectionState newState) {
    if (isClosed_) return;
    
    if (!isValidStateTransition(state_, newState))
        return;
    state_ = newState;
    switch (state_) {
        case signaling::ConnectionState::Connected:
            postToDispatcher([](PeerSession& self){
                self.signalStream_->sendConnectionState(self.state_, self.sessionId_);
            });
            break;
        case signaling::ConnectionState::Failed:
        case signaling::ConnectionState::Closed:
            onSessionTerminated();
            break;
        default:
            break;
    }
}

bool PeerSession::isValidStateTransition(signaling::ConnectionState from, signaling::ConnectionState to) {
    switch (from) {
        case signaling::ConnectionState::New:
            return to == signaling::ConnectionState::Connecting || to == signaling::ConnectionState::Closed;
        case signaling::ConnectionState::Connecting:
            return to == signaling::ConnectionState::Connected || to == signaling::ConnectionState::Failed;
        case signaling::ConnectionState::Connected:
            return to == signaling::ConnectionState::Disconnected || to == signaling::ConnectionState::Closed;
        default:
            return false;
    }
}


auto PeerSession::attachSignalStream(const std::shared_ptr<signaling::ISignalingSink>& signalStream) -> std::expected<void, AttachSignalStreamError> {
    std::lock_guard<std::mutex> lock(mutex_);
    if (isClosed_)
        return std::unexpected(AttachSignalStreamError::SessionClosed);
    
    if (!signalStream) 
        return std::unexpected(AttachSignalStreamError::InvalidSignalStream);
    
    signalStream_ = signalStream;
    return {};
}

auto PeerSession::setRemoteDescription(const signaling::Description& desc) -> std::expected<void, SetRemoteDescriptionError> {
    std::lock_guard<std::mutex> lock(mutex_);
    if (isClosed_)
        return std::unexpected(SetRemoteDescriptionError::SessionClosed);
    
    if (!rtcPeer_)
        return std::unexpected(SetRemoteDescriptionError::PeerConnectionNotStarted);

    if (desc.sdp.empty())
        return std::unexpected(SetRemoteDescriptionError::InvalidDescription);
    
    rtcPeer_->setRemoteDescription(desc);
    return {};
}

auto PeerSession::addIceCandidate(const signaling::IceCandidate& candidate) -> std::expected<void, IceCandidateError> {
    std::lock_guard<std::mutex> lock(mutex_);
    if (isClosed_)
        return std::unexpected(IceCandidateError::SessionClosed);
    
    if (!rtcPeer_)
        return std::unexpected(IceCandidateError::RemoteDescriptionNotSet);

    if (candidate.candidate.empty() || candidate.sdpMid.empty() || candidate.sdpMLineIndex < 0)
        return std::unexpected(IceCandidateError::InvalidCandidate);
    
    rtcPeer_->addRemoteIceCandidate(candidate);
    return {};
}

void PeerSession::close() {
    std::lock_guard<std::mutex> lock(mutex_);
    rtcPeer_.reset();
    signalStream_.reset();
    isClosed_ = true;
}