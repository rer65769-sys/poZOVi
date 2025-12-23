#pragma once
#include <string>

namespace webrtc::signaling {

enum class ConnectionState {
    New,
    Connecting,
    Connected,
    Disconnected,
    Failed,
    Closed
};

enum class MessageType { 
    Unspec,
    Offer,
    Answer,
    Pranswer,
    Rollback
};

struct Description {
    std::string sdp;
    MessageType type;
};

struct IceCandidate {
    std::string candidate;
    std::string sdpMid;
    int sdpMLineIndex;
};

}