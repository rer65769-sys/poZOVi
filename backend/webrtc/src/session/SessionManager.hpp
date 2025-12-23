#pragma once

#include <mutex>
#include <unordered_map>
#include <memory>
#include <expected>
#include <optional>
#include "../dispatcher/Dispatcher.hpp"

namespace webrtc::session {

class PeerSession;

enum class CreateSessionError {
    AlreadyExists,
    InvalidSessionId,
    ResourceUnavailable
};

class SessionManager {
public:
    SessionManager(dispatch::Dispatcher& dispatcher) : dispatcher_(dispatcher) {};
    ~SessionManager() = default;

    auto createSession(uint32_t sessionId) -> std::expected<void, CreateSessionError>;
    auto getSession(uint32_t sessionId) -> std::expected<std::shared_ptr<PeerSession>, nullptr_t>;
    void closeSession(uint32_t sessionId);

private:
    std::mutex mutex_;
    std::unordered_map<uint32_t, std::shared_ptr<PeerSession>> sessions_;
    dispatch::Dispatcher& dispatcher_;
};

}