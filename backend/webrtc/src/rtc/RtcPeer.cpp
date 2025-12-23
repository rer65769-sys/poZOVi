#include "RtcPeer.hpp"

using namespace webrtc::rtc;


void RtcPeer::populateMidToIndexMap(const ::rtc::Description& desc) {
    midToIndexMap_.clear(); // With new description, the old sdp becomes invalid, therefore we clear the map
    for (int i = 0; i < desc.mediaCount(); ++i) {
        const auto& mediaVar = desc.media(i);
        std::visit([this, i](auto* media) {
            midToIndexMap_[media->mid()] = i;
        }, mediaVar);
    }
}

void RtcPeer::start(::rtc::Configuration config) {
    config_ = config;
    peerConnection_ = std::make_unique<::rtc::PeerConnection>(config_);

    peerConnection_->onLocalDescription([this](::rtc::Description sdp) {
        populateMidToIndexMap(sdp);
        signaling::Description localDesc;
        localDesc.sdp = sdp;
        localDesc.type = static_cast<signaling::MessageType>(sdp.type());
        
        if (onLocalDescription)
            onLocalDescription(localDesc);
    });

    peerConnection_->onLocalCandidate([this](::rtc::Candidate candidate) {
        signaling::IceCandidate iceCandidate;
        iceCandidate.candidate = candidate.candidate();
        iceCandidate.sdpMid = candidate.mid();
        iceCandidate.sdpMLineIndex = midToIndexMap_.at(candidate.mid());

        if (onIceCandidate)
            onIceCandidate(iceCandidate);
    });

    peerConnection_->onStateChange([this](::rtc::PeerConnection::State state) {
        signaling::ConnectionState connectionState = static_cast<signaling::ConnectionState>(state);
        if (onConnectionStateChange)
            onConnectionStateChange(connectionState);
    });
}

void RtcPeer::close() {
    if (closed_.exchange(true)) return;
    midToIndexMap_.clear();
    if (peerConnection_) {
        peerConnection_->close();
        peerConnection_.reset();
    }
}

void RtcPeer::setRemoteDescription(const signaling::Description& desc) {
    if (peerConnection_) {
        const auto sdp = desc.sdp;
        const auto type = static_cast<::rtc::Description::Type>(desc.type);
        ::rtc::Description description(sdp, type);
        peerConnection_->setRemoteDescription(description);
    }
}

void RtcPeer::addRemoteIceCandidate(const signaling::IceCandidate& candidate) {
    if (peerConnection_) {
        ::rtc::Candidate rtcCandidate(candidate.candidate, candidate.sdpMid);
        peerConnection_->addRemoteCandidate(rtcCandidate);
    }
}

