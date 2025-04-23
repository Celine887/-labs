#include <gtest/gtest.h>
#include "../lib/scheduler.h"
#include <cmath>
#include <string>
#include <limits>

struct TestStruct {
    float multiplyByValue(float input) const {
        return input * value;
    }
    
    float value;
};

TEST(SchedulerTest, BasicTask) {
    TTaskScheduler scheduler;
    
    auto id = scheduler.add([]() { return 42; });
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id), 42);
}

TEST(SchedulerTest, TaskWithOneArg) {
    TTaskScheduler scheduler;
    
    auto id = scheduler.add([](int x) { return x * 2; }, 21);
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id), 42);
}

TEST(SchedulerTest, TaskWithTwoArgs) {
    TTaskScheduler scheduler;
    
    auto id = scheduler.add([](int x, int y) { return x + y; }, 20, 22);
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id), 42);
}

TEST(SchedulerTest, MemberFunction) {
    TTaskScheduler scheduler;
    TestStruct test{2.0f};
    
    auto id = scheduler.add(&TestStruct::multiplyByValue, test, 21.0f);
    scheduler.executeAll();
    
    EXPECT_FLOAT_EQ(scheduler.getResult<float>(id), 42.0f);
}

TEST(SchedulerTest, TaskDependencies) {
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([]() { return 10; });
    auto id2 = scheduler.add([](int x) { return x * 2; }, scheduler.getFutureResult<int>(id1));
    auto id3 = scheduler.add([](int x, int y) { return x + y; }, 
                             scheduler.getFutureResult<int>(id2), 
                             22);
    
    EXPECT_EQ(scheduler.getResult<int>(id3), 42);
}

TEST(SchedulerTest, ZeroValueOperations) {
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([]() { return 0; });
    auto id2 = scheduler.add([](int x) { return x + 10; }, scheduler.getFutureResult<int>(id1));
    auto id3 = scheduler.add([](int x, int y) { return x * y; }, 
                            scheduler.getFutureResult<int>(id2), 
                            0);
    
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id1), 0);
    EXPECT_EQ(scheduler.getResult<int>(id2), 10);
    EXPECT_EQ(scheduler.getResult<int>(id3), 0);
}

TEST(SchedulerTest, NegativeValues) {
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([]() { return -5; });
    auto id2 = scheduler.add([](int x) { return x * -2; }, scheduler.getFutureResult<int>(id1));
    
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id1), -5);
    EXPECT_EQ(scheduler.getResult<int>(id2), 10);
}

TEST(SchedulerTest, LongTaskChain) {
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([]() { return 1; });
    auto id2 = scheduler.add([](int x) { return x + 1; }, scheduler.getFutureResult<int>(id1));
    auto id3 = scheduler.add([](int x) { return x + 1; }, scheduler.getFutureResult<int>(id2));
    auto id4 = scheduler.add([](int x) { return x + 1; }, scheduler.getFutureResult<int>(id3));
    auto id5 = scheduler.add([](int x) { return x + 1; }, scheduler.getFutureResult<int>(id4));
    auto id6 = scheduler.add([](int x) { return x + 1; }, scheduler.getFutureResult<int>(id5));
    
    scheduler.executeAll();
    
    EXPECT_EQ(scheduler.getResult<int>(id6), 6);
}

TEST(SchedulerTest, QuadraticEquation) {
    float a = 1;
    float b = -3;
    float c = 2;
    
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([](float a, float c) { return -4 * a * c; }, a, c);
    
    auto id2 = scheduler.add([](float b, float v) { return b * b + v; }, b, scheduler.getFutureResult<float>(id1));
    
    auto id3 = scheduler.add([](float b, float d) { return -b + std::sqrt(d); }, b, scheduler.getFutureResult<float>(id2));
    auto id5 = scheduler.add([](float a, float v) { return v / (2 * a); }, a, scheduler.getFutureResult<float>(id3));
    
    auto id4 = scheduler.add([](float b, float d) { return -b - std::sqrt(d); }, b, scheduler.getFutureResult<float>(id2));
    auto id6 = scheduler.add([](float a, float v) { return v / (2 * a); }, a, scheduler.getFutureResult<float>(id4));
    
    scheduler.executeAll();
    
    EXPECT_FLOAT_EQ(scheduler.getResult<float>(id5), 2.0f);
    EXPECT_FLOAT_EQ(scheduler.getResult<float>(id6), 1.0f);
}

TEST(SchedulerTest, FloatingPointEdgeCases) {
    TTaskScheduler scheduler;
    
    auto id1 = scheduler.add([]() { return 1e-10f; });
    auto id2 = scheduler.add([](float x) { return x * 1e10f; }, scheduler.getFutureResult<float>(id1));
    
    auto id3 = scheduler.add([]() { return std::numeric_limits<float>::infinity(); }); // Infinity
    auto id4 = scheduler.add([](float x) { return std::isfinite(x) ? 42.0f : -1.0f; }, 
                           scheduler.getFutureResult<float>(id3));
    
    scheduler.executeAll();
    
    EXPECT_NEAR(scheduler.getResult<float>(id2), 1.0f, 1e-5f);
    EXPECT_EQ(scheduler.getResult<float>(id4), -1.0f);
}

int main(int argc, char** argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
} 