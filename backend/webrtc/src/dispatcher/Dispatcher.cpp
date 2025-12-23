#include "Dispatcher.hpp"

using namespace webrtc::dispatch;

void Dispatcher::start() {
    if (running_.exchange(true)) return;
    worker_ = std::thread([this]() {
        while (running_ && !taskQueue_.empty()) {
            std::function<void()> task;
            {
                std::unique_lock<std::mutex> lock(mutex_);
                cv_.wait(lock, [this]() { return !taskQueue_.empty() || !running_; });
                task = taskQueue_.front();
                taskQueue_.pop();
            }
            task();
        }
    });
}


void Dispatcher::post(const std::function<void()>& task) {
    {
        std::lock_guard<std::mutex> lock(mutex_);
        taskQueue_.push(task);
    }
    cv_.notify_one();
}

Dispatcher::~Dispatcher() {
    if (!running_.exchange(false)) return;
    cv_.notify_all();
    if (worker_.joinable())
        worker_.join();
}