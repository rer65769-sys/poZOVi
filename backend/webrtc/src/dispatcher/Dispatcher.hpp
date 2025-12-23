#pragma once
#include <queue>
#include <mutex>
#include <functional>
#include <condition_variable>
#include <thread>
#include <atomic>

namespace webrtc::dispatch {

class Dispatcher {
public:
    ~Dispatcher();
    void start();
    void post(const std::function<void()>& task);
    
private:
    std::mutex mutex_;
    std::queue<std::function<void()>> taskQueue_;
    std::condition_variable cv_;
    std::thread worker_;
    std::atomic<bool> running_;
};

}