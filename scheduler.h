#include <vector>
#include <memory>

struct TypeId {
    const void* address;
    
    bool operator==(const TypeId& other) const {
        return address == other.address;
    }
    
    bool operator!=(const TypeId& other) const {
        return address != other.address;
    }
};

template<typename T>
TypeId get_type_id() {
    static char dummy;
    return TypeId{&dummy};
}

class Task {
public:
    virtual ~Task() = default;
    virtual void compute() = 0;
    virtual bool hasResult() const = 0;
    virtual TypeId resultType() const = 0;
};

template<typename ResultType>
class TypedTask : public Task {
protected:
    ResultType _result;
    bool _computed = false;
public:
    TypeId resultType() const override {
        return get_type_id<ResultType>();
    }
    
    bool hasResult() const override {
        return _computed;
    }
    
    ResultType getResult() {
        if (!_computed) {
            compute();
        }
        return _result;
    }
};

template<typename T>
struct FutureResult {
    Task* _task;
    explicit FutureResult(Task* task) : _task(task) {}
    
    T getValue() const {
        if (_task->resultType() != get_type_id<T>()) {
            return *static_cast<T*>(nullptr);
        }
        auto* typedTask = static_cast<TypedTask<T>*>(_task);
        return typedTask->getResult();
    }
};

template<typename T>
class TaskArg {
    T _argument;
public:
    TaskArg() = default;
    explicit TaskArg(const T& arg) : _argument(arg) {}
    T getArgument() { return _argument; }
};

template<typename T>
class TaskArg<FutureResult<T>> {
    FutureResult<T> _argument;
public:
    TaskArg() = default;
    explicit TaskArg(const FutureResult<T>& arg) : _argument(arg) {}
    T getArgument() { 
        return _argument.getValue();
    }
};

template<typename Function, typename ResultType>
class ZeroArgumentTask : public TypedTask<ResultType> {
    Function _func;
public:
    explicit ZeroArgumentTask(const Function& func) : _func(func) {}
    void compute() override {
        this->_result = _func();
        this->_computed = true;
    }
};

template<typename Function, typename Arg, typename ResultType>
class OneArgumentTask : public TypedTask<ResultType> {
    Function _func;
    TaskArg<Arg> _arg;
public:
    OneArgumentTask(const Function& func, const Arg& arg) : _func(func), _arg(arg) {}
    void compute() override {
        this->_result = _func(_arg.getArgument());
        this->_computed = true;
    }
};

template<typename Function, typename A, typename B, typename ResultType>
class TwoArgumentTask : public TypedTask<ResultType> {
    Function _func;
    TaskArg<A> _first;
    TaskArg<B> _second;
public:
    TwoArgumentTask(const Function& func, const A& a, const B& b)
        : _func(func), _first(a), _second(b) {}
    void compute() override {
        this->_result = _func(_first.getArgument(), _second.getArgument());
        this->_computed = true;
    }
};

template<typename Class, typename Return, typename Arg>
class MemberFunctionTask : public TypedTask<Return> {
    Return (Class::*_func)(Arg) const;
    TaskArg<Class> _obj;
    TaskArg<Arg> _arg;
public:
    MemberFunctionTask(Return (Class::*func)(Arg) const, const Class& obj, const Arg& arg)
        : _func(func), _obj(obj), _arg(arg) {}
    
    void compute() override {
        Class obj = _obj.getArgument();
        this->_result = (obj.*_func)(_arg.getArgument());
        this->_computed = true;
    }
};

template<typename Class, typename Return, typename Arg>
class MemberFunctionWithFutureTask : public TypedTask<Return> {
    Return (Class::*_func)(Arg) const;
    TaskArg<Class> _obj;
    TaskArg<FutureResult<Arg>> _arg;
public:
    MemberFunctionWithFutureTask(Return (Class::*func)(Arg) const, const Class& obj, const FutureResult<Arg>& arg)
        : _func(func), _obj(obj), _arg(arg) {}
    
    void compute() override {
        Class obj = _obj.getArgument();
        this->_result = (obj.*_func)(_arg.getArgument());
        this->_computed = true;
    }
};

class TaskOrder {
    unsigned _id;
public:
    explicit TaskOrder(unsigned id) : _id(id) {}
    unsigned id() const { return _id; }
};

 template <typename T>
struct is_future_result {
    static constexpr bool value = false;
};

template <typename T>
struct is_future_result<FutureResult<T>> {
    static constexpr bool value = true;
};

template <typename T>
constexpr bool is_future_result_v = is_future_result<T>::value;

template <typename T, bool is_future = is_future_result<T>::value>
struct param_value_type { using type = T; };

template <typename T>
struct param_value_type<T, true> { using type = typename T::value_type; };

template<bool B, typename T = void>
struct enable_if {};

template<typename T>
struct enable_if<true, T> { using type = T; };

template<bool B, typename T = void>
using enable_if_t = typename enable_if<B, T>::type;

class TTaskScheduler {
    std::vector<std::unique_ptr<Task>> _tasks;
public:
    TTaskScheduler() = default;
    
    template<typename Return, typename Class, typename Arg>
    TaskOrder add(Return (Class::*func)(Arg) const, const Class& obj, const Arg& arg) {
        _tasks.emplace_back(new MemberFunctionTask<Class, Return, Arg>(func, obj, arg));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Return, typename Class, typename Arg>
    TaskOrder add(Return (Class::*func)(Arg) const, const Class& obj, const FutureResult<Arg>& futureArg) {
        _tasks.emplace_back(new MemberFunctionWithFutureTask<Class, Return, Arg>(func, obj, futureArg));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func>
    TaskOrder add(const Func& func) {
        using ResultType = decltype(func());
        _tasks.emplace_back(new ZeroArgumentTask<Func, ResultType>(func));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename Arg>
    enable_if_t<!is_future_result<Arg>::value, TaskOrder>
    add(const Func& func, const Arg& arg) {
        using ResultType = decltype(func(arg));
        _tasks.emplace_back(new OneArgumentTask<Func, Arg, ResultType>(func, arg));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename T>
    enable_if_t<is_future_result<FutureResult<T>>::value, TaskOrder>
    add(const Func& func, const FutureResult<T>& arg) {
        using ResultType = decltype(func(T()));
        _tasks.emplace_back(new OneArgumentTask<Func, FutureResult<T>, ResultType>(func, arg));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename A, typename B>
    TaskOrder addWithTwoArgs(const Func& func, const A& a, const B& b, 
                        typename enable_if<!is_future_result<A>::value && 
                                          !is_future_result<B>::value>::type* = nullptr) {
        using ResultType = decltype(func(a, b));
        _tasks.emplace_back(new TwoArgumentTask<Func, A, B, ResultType>(func, a, b));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename T, typename B>
    TaskOrder addWithTwoArgs(const Func& func, const FutureResult<T>& a, const B& b,
                        typename enable_if<!is_future_result<B>::value>::type* = nullptr) {
        using ResultType = decltype(func(T(), b));
        _tasks.emplace_back(new TwoArgumentTask<Func, FutureResult<T>, B, ResultType>(func, a, b));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename A, typename T>
    TaskOrder addWithTwoArgs(const Func& func, const A& a, const FutureResult<T>& b,
                        typename enable_if<!is_future_result<A>::value>::type* = nullptr) {
        using ResultType = decltype(func(a, T()));
        _tasks.emplace_back(new TwoArgumentTask<Func, A, FutureResult<T>, ResultType>(func, a, b));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename T1, typename T2>
    TaskOrder addWithTwoArgs(const Func& func, const FutureResult<T1>& a, const FutureResult<T2>& b) {
        using ResultType = decltype(func(T1(), T2()));
        _tasks.emplace_back(new TwoArgumentTask<Func, FutureResult<T1>, FutureResult<T2>, ResultType>(func, a, b));
        return TaskOrder(_tasks.size() - 1);
    }
    
    template<typename Func, typename A, typename B>
    TaskOrder add(const Func& func, const A& a, const B& b) {
        return addWithTwoArgs(func, a, b);
    }
    
    template<typename T>
    FutureResult<T> getFutureResult(const TaskOrder& order) {
        return FutureResult<T>(_tasks[order.id()].get());
    }
    
    template<typename T>
    T getResult(const TaskOrder& order) {
        Task* task = _tasks[order.id()].get();
        if (task->resultType() != get_type_id<T>()) {
            return T();
        }
        auto* typedTask = static_cast<TypedTask<T>*>(task);
        return typedTask->getResult();
    }
    
    void executeAll() {
        for (auto& task : _tasks) {
            task->compute();
        }
    }
};