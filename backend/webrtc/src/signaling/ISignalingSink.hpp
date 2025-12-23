#pragma once

#include "SignalingTypes.hpp"

namespace webrtc::signaling {
    
class ISignalingSink {
public:
    virtual ~ISignalingSink() = default;

    virtual void sendLocalDescription(const Description& desc, const uint32_t id) = 0;
    virtual void sendIceCandidate(const IceCandidate& candidate, const uint32_t id) = 0;
    virtual void sendConnectionState(const ConnectionState state, const uint32_t id) = 0;
};

}