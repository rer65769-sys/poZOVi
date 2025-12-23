#include "SignalStream.hpp"
#include "../signaling/SignalingTypes.hpp"
#include "../session/PeerSession.hpp"

using namespace webrtc::signaling;

void SignalStream::start() {
    while (!closed_) {
        webrtc::SignalingMessage message;
        if (!stream_->Read(&message) || closed_) 
            break;

        switch (message.message_case()) {
            case webrtc::SignalingMessage::kDescription:
                handleRemoteDescription(message);
                break;
            case webrtc::SignalingMessage::kIceCandidate:
                handleIceCandidate(message);
                break;
            case webrtc::SignalingMessage::kState:
                handleStateChange(message);
                break;
            default:
                break;
        }
    }
}

template<typename Fn>
void SignalStream::withSession(uint32_t sessionId, Fn&& fn) {
    auto weakSelf = weak_from_this();

    dispatcher_.post([weakSelf, sessionId, fn = std::move(fn)]() mutable {
        if (auto self = weakSelf.lock()) {
            if (auto s = self->sessionManager_.getSession(sessionId)) 
                fn(*s);
        }
    });
}

void SignalStream::handleRemoteDescription(const webrtc::SignalingMessage& message) {
    withSession(message.session_id(),
        [message](const std::shared_ptr<session::PeerSession>& session) {
            signaling::Description desc{
                .sdp  = message.description().sdp(),
                .type = static_cast<signaling::MessageType>(message.description().type())
            };

            session->setRemoteDescription(desc);
        }
    );
}

void SignalStream::handleIceCandidate(const webrtc::SignalingMessage& message) {
    withSession(message.session_id(),
        [message](const std::shared_ptr<session::PeerSession>& session) {
            signaling::IceCandidate candidate{
                .candidate = message.ice_candidate().candidate(),
                .sdpMid = message.ice_candidate().sdpmid(),
                .sdpMLineIndex = message.ice_candidate().sdpmlineindex()
            };

            session->addIceCandidate(candidate);
        }
    );
}

void SignalStream::createPeerSession(uint32_t id) {
    auto weakSelf = weak_from_this();
    dispatcher_.post([weakSelf, id]() {
        if (auto self = weakSelf.lock())
            self->sessionManager_.createSession(id);
    });
}

void SignalStream::handleStateChange(const webrtc::SignalingMessage& message) {
    if (message.state().state() == webrtc::ConnectionState::NEW)
        return createPeerSession(message.session_id());
    
    auto state = static_cast<signaling::ConnectionState>(message.state().state());
    withSession(message.session_id(),
        [message](const std::shared_ptr<session::PeerSession>& session) {
            auto state = static_cast<ConnectionState>(message.state().state());
            session->handleConnectionStateChange(state);
        }
    );
}

void SignalStream::send(const webrtc::SignalingMessage& message) {
    if (closed_.load()) return;
    if (!stream_->Write(message)) 
        closed_.store(true);
}

void SignalStream::close() {
    if (closed_.exchange(true)) return;
}

void SignalStream::sendLocalDescription(const Description& desc, const uint32_t id) {
    webrtc::SignalingMessage message;
    message.set_session_id(id);

    auto* protoDesc = message.mutable_description();
    protoDesc->set_sdp(desc.sdp);
    protoDesc->set_type(
        static_cast<webrtc::SessionDescription::Type>(desc.type)
    );

    send(message);
}

void SignalStream::sendIceCandidate(const IceCandidate& candidate, const uint32_t id) {
    webrtc::SignalingMessage message;
    message.set_session_id(id);

    auto* protoDesc = message.mutable_ice_candidate();
    protoDesc->set_candidate(candidate.candidate);
    protoDesc->set_sdpmid(candidate.sdpMid);
    protoDesc->set_sdpmlineindex(candidate.sdpMLineIndex);

    send(message);
}

void SignalStream::sendConnectionState(const ConnectionState state, const uint32_t id) {
    webrtc::SignalingMessage message;
    message.set_session_id(id);

    auto* protocDesc = message.mutable_state();
    protocDesc->set_state(
        static_cast<webrtc::ConnectionState::State>(state)
    );
    
    send(message);
}