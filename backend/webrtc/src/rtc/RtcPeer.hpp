#pragma once
#include <functional>
#include <memory>
#include <atomic>
#include <unordered_map>
#include <rtc/rtc.hpp>
#include "../signaling/SignalingTypes.hpp"

namespace webrtc::rtc {

class RtcPeer {
public:
    RtcPeer() = default;
    ~RtcPeer() { close(); };
    void start();
    void close();
    void setRemoteDescription(const signaling::Description& desc);
    void addRemoteIceCandidate(const signaling::IceCandidate& candidate);

    std::function<void(const signaling::Description&)> onLocalDescription;
    std::function<void(const signaling::IceCandidate&)> onIceCandidate;
    std::function<void(signaling::ConnectionState)> onConnectionStateChange;
    
private:
    ::rtc::Configuration config_;
    std::unique_ptr<::rtc::PeerConnection> peerConnection_;
    std::atomic<bool> closed_{false};
    std::unordered_map<std::string, int> midToIndexMap_;

    void populateMidToIndexMap(const ::rtc::Description& desc);
};

}