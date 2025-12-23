#include "SessionManager.hpp"
#include "PeerSession.hpp"

using namespace webrtc::session;

auto SessionManager::createSession(uint32_t sessionId) -> std::expected<void, CreateSessionError> {
    std::lock_guard<std::mutex> lock(mutex_);
    if (sessions_.find(sessionId) != sessions_.end())
        return std::unexpected(CreateSessionError::AlreadyExists);

    std::function<void()> onSessionTerminated = [this, sessionId]() {
        this->closeSession(sessionId);
    };
    
    sessions_.try_emplace(sessionId, std::make_shared<PeerSession>(dispatcher_, onSessionTerminated));
    return {};
}

auto SessionManager::getSession(uint32_t sessionId) -> std::expected<std::shared_ptr<PeerSession>, nullptr_t> {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = sessions_.find(sessionId);
    if (it != sessions_.end())
        return it->second;
    
    return std::unexpected(nullptr);
}

void SessionManager::closeSession(uint32_t sessionId) {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = sessions_.find(sessionId);
    if (it == sessions_.end()) return;
    it->second->close();
    sessions_.erase(sessionId);
}
