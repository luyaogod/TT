var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __esm = (fn, res) => function __init() {
  return fn && (res = (0, fn[__getOwnPropNames(fn)[0]])(fn = 0)), res;
};
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));

// ../node_modules/zustand/esm/vanilla.mjs
var createStoreImpl, createStore;
var init_vanilla = __esm({
  "../node_modules/zustand/esm/vanilla.mjs"() {
    createStoreImpl = (createState) => {
      let state;
      const listeners = /* @__PURE__ */ new Set();
      const setState = (partial, replace) => {
        const nextState = typeof partial === "function" ? partial(state) : partial;
        if (!Object.is(nextState, state)) {
          const previousState = state;
          state = (replace != null ? replace : typeof nextState !== "object" || nextState === null) ? nextState : Object.assign({}, state, nextState);
          listeners.forEach((listener) => listener(state, previousState));
        }
      };
      const getState = () => state;
      const getInitialState = () => initialState;
      const subscribe = (listener) => {
        listeners.add(listener);
        return () => listeners.delete(listener);
      };
      const api2 = { setState, getState, getInitialState, subscribe };
      const initialState = state = createState(setState, getState, api2);
      return api2;
    };
    createStore = ((createState) => createState ? createStoreImpl(createState) : createStoreImpl);
  }
});

// ../node_modules/react/cjs/react.production.min.js
var require_react_production_min = __commonJS({
  "../node_modules/react/cjs/react.production.min.js"(exports) {
    "use strict";
    var l = Symbol.for("react.element");
    var n = Symbol.for("react.portal");
    var p4 = Symbol.for("react.fragment");
    var q = Symbol.for("react.strict_mode");
    var r = Symbol.for("react.profiler");
    var t = Symbol.for("react.provider");
    var u = Symbol.for("react.context");
    var v = Symbol.for("react.forward_ref");
    var w = Symbol.for("react.suspense");
    var x = Symbol.for("react.memo");
    var y = Symbol.for("react.lazy");
    var z = Symbol.iterator;
    function A(a) {
      if (null === a || "object" !== typeof a) return null;
      a = z && a[z] || a["@@iterator"];
      return "function" === typeof a ? a : null;
    }
    var B = { isMounted: function() {
      return false;
    }, enqueueForceUpdate: function() {
    }, enqueueReplaceState: function() {
    }, enqueueSetState: function() {
    } };
    var C = Object.assign;
    var D = {};
    function E(a, b, e) {
      this.props = a;
      this.context = b;
      this.refs = D;
      this.updater = e || B;
    }
    E.prototype.isReactComponent = {};
    E.prototype.setState = function(a, b) {
      if ("object" !== typeof a && "function" !== typeof a && null != a) throw Error("setState(...): takes an object of state variables to update or a function which returns an object of state variables.");
      this.updater.enqueueSetState(this, a, b, "setState");
    };
    E.prototype.forceUpdate = function(a) {
      this.updater.enqueueForceUpdate(this, a, "forceUpdate");
    };
    function F() {
    }
    F.prototype = E.prototype;
    function G(a, b, e) {
      this.props = a;
      this.context = b;
      this.refs = D;
      this.updater = e || B;
    }
    var H = G.prototype = new F();
    H.constructor = G;
    C(H, E.prototype);
    H.isPureReactComponent = true;
    var I = Array.isArray;
    var J = Object.prototype.hasOwnProperty;
    var K = { current: null };
    var L = { key: true, ref: true, __self: true, __source: true };
    function M(a, b, e) {
      var d, c = {}, k = null, h = null;
      if (null != b) for (d in void 0 !== b.ref && (h = b.ref), void 0 !== b.key && (k = "" + b.key), b) J.call(b, d) && !L.hasOwnProperty(d) && (c[d] = b[d]);
      var g = arguments.length - 2;
      if (1 === g) c.children = e;
      else if (1 < g) {
        for (var f = Array(g), m = 0; m < g; m++) f[m] = arguments[m + 2];
        c.children = f;
      }
      if (a && a.defaultProps) for (d in g = a.defaultProps, g) void 0 === c[d] && (c[d] = g[d]);
      return { $$typeof: l, type: a, key: k, ref: h, props: c, _owner: K.current };
    }
    function N(a, b) {
      return { $$typeof: l, type: a.type, key: b, ref: a.ref, props: a.props, _owner: a._owner };
    }
    function O(a) {
      return "object" === typeof a && null !== a && a.$$typeof === l;
    }
    function escape(a) {
      var b = { "=": "=0", ":": "=2" };
      return "$" + a.replace(/[=:]/g, function(a2) {
        return b[a2];
      });
    }
    var P = /\/+/g;
    function Q(a, b) {
      return "object" === typeof a && null !== a && null != a.key ? escape("" + a.key) : b.toString(36);
    }
    function R(a, b, e, d, c) {
      var k = typeof a;
      if ("undefined" === k || "boolean" === k) a = null;
      var h = false;
      if (null === a) h = true;
      else switch (k) {
        case "string":
        case "number":
          h = true;
          break;
        case "object":
          switch (a.$$typeof) {
            case l:
            case n:
              h = true;
          }
      }
      if (h) return h = a, c = c(h), a = "" === d ? "." + Q(h, 0) : d, I(c) ? (e = "", null != a && (e = a.replace(P, "$&/") + "/"), R(c, b, e, "", function(a2) {
        return a2;
      })) : null != c && (O(c) && (c = N(c, e + (!c.key || h && h.key === c.key ? "" : ("" + c.key).replace(P, "$&/") + "/") + a)), b.push(c)), 1;
      h = 0;
      d = "" === d ? "." : d + ":";
      if (I(a)) for (var g = 0; g < a.length; g++) {
        k = a[g];
        var f = d + Q(k, g);
        h += R(k, b, e, f, c);
      }
      else if (f = A(a), "function" === typeof f) for (a = f.call(a), g = 0; !(k = a.next()).done; ) k = k.value, f = d + Q(k, g++), h += R(k, b, e, f, c);
      else if ("object" === k) throw b = String(a), Error("Objects are not valid as a React child (found: " + ("[object Object]" === b ? "object with keys {" + Object.keys(a).join(", ") + "}" : b) + "). If you meant to render a collection of children, use an array instead.");
      return h;
    }
    function S2(a, b, e) {
      if (null == a) return a;
      var d = [], c = 0;
      R(a, d, "", "", function(a2) {
        return b.call(e, a2, c++);
      });
      return d;
    }
    function T(a) {
      if (-1 === a._status) {
        var b = a._result;
        b = b();
        b.then(function(b2) {
          if (0 === a._status || -1 === a._status) a._status = 1, a._result = b2;
        }, function(b2) {
          if (0 === a._status || -1 === a._status) a._status = 2, a._result = b2;
        });
        -1 === a._status && (a._status = 0, a._result = b);
      }
      if (1 === a._status) return a._result.default;
      throw a._result;
    }
    var U = { current: null };
    var V = { transition: null };
    var W = { ReactCurrentDispatcher: U, ReactCurrentBatchConfig: V, ReactCurrentOwner: K };
    function X() {
      throw Error("act(...) is not supported in production builds of React.");
    }
    exports.Children = { map: S2, forEach: function(a, b, e) {
      S2(a, function() {
        b.apply(this, arguments);
      }, e);
    }, count: function(a) {
      var b = 0;
      S2(a, function() {
        b++;
      });
      return b;
    }, toArray: function(a) {
      return S2(a, function(a2) {
        return a2;
      }) || [];
    }, only: function(a) {
      if (!O(a)) throw Error("React.Children.only expected to receive a single React element child.");
      return a;
    } };
    exports.Component = E;
    exports.Fragment = p4;
    exports.Profiler = r;
    exports.PureComponent = G;
    exports.StrictMode = q;
    exports.Suspense = w;
    exports.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED = W;
    exports.act = X;
    exports.cloneElement = function(a, b, e) {
      if (null === a || void 0 === a) throw Error("React.cloneElement(...): The argument must be a React element, but you passed " + a + ".");
      var d = C({}, a.props), c = a.key, k = a.ref, h = a._owner;
      if (null != b) {
        void 0 !== b.ref && (k = b.ref, h = K.current);
        void 0 !== b.key && (c = "" + b.key);
        if (a.type && a.type.defaultProps) var g = a.type.defaultProps;
        for (f in b) J.call(b, f) && !L.hasOwnProperty(f) && (d[f] = void 0 === b[f] && void 0 !== g ? g[f] : b[f]);
      }
      var f = arguments.length - 2;
      if (1 === f) d.children = e;
      else if (1 < f) {
        g = Array(f);
        for (var m = 0; m < f; m++) g[m] = arguments[m + 2];
        d.children = g;
      }
      return { $$typeof: l, type: a.type, key: c, ref: k, props: d, _owner: h };
    };
    exports.createContext = function(a) {
      a = { $$typeof: u, _currentValue: a, _currentValue2: a, _threadCount: 0, Provider: null, Consumer: null, _defaultValue: null, _globalName: null };
      a.Provider = { $$typeof: t, _context: a };
      return a.Consumer = a;
    };
    exports.createElement = M;
    exports.createFactory = function(a) {
      var b = M.bind(null, a);
      b.type = a;
      return b;
    };
    exports.createRef = function() {
      return { current: null };
    };
    exports.forwardRef = function(a) {
      return { $$typeof: v, render: a };
    };
    exports.isValidElement = O;
    exports.lazy = function(a) {
      return { $$typeof: y, _payload: { _status: -1, _result: a }, _init: T };
    };
    exports.memo = function(a, b) {
      return { $$typeof: x, type: a, compare: void 0 === b ? null : b };
    };
    exports.startTransition = function(a) {
      var b = V.transition;
      V.transition = {};
      try {
        a();
      } finally {
        V.transition = b;
      }
    };
    exports.unstable_act = X;
    exports.useCallback = function(a, b) {
      return U.current.useCallback(a, b);
    };
    exports.useContext = function(a) {
      return U.current.useContext(a);
    };
    exports.useDebugValue = function() {
    };
    exports.useDeferredValue = function(a) {
      return U.current.useDeferredValue(a);
    };
    exports.useEffect = function(a, b) {
      return U.current.useEffect(a, b);
    };
    exports.useId = function() {
      return U.current.useId();
    };
    exports.useImperativeHandle = function(a, b, e) {
      return U.current.useImperativeHandle(a, b, e);
    };
    exports.useInsertionEffect = function(a, b) {
      return U.current.useInsertionEffect(a, b);
    };
    exports.useLayoutEffect = function(a, b) {
      return U.current.useLayoutEffect(a, b);
    };
    exports.useMemo = function(a, b) {
      return U.current.useMemo(a, b);
    };
    exports.useReducer = function(a, b, e) {
      return U.current.useReducer(a, b, e);
    };
    exports.useRef = function(a) {
      return U.current.useRef(a);
    };
    exports.useState = function(a) {
      return U.current.useState(a);
    };
    exports.useSyncExternalStore = function(a, b, e) {
      return U.current.useSyncExternalStore(a, b, e);
    };
    exports.useTransition = function() {
      return U.current.useTransition();
    };
    exports.version = "18.3.1";
  }
});

// ../node_modules/react/cjs/react.development.js
var require_react_development = __commonJS({
  "../node_modules/react/cjs/react.development.js"(exports, module) {
    "use strict";
    if (process.env.NODE_ENV !== "production") {
      (function() {
        "use strict";
        if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ !== "undefined" && typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.registerInternalModuleStart === "function") {
          __REACT_DEVTOOLS_GLOBAL_HOOK__.registerInternalModuleStart(new Error());
        }
        var ReactVersion = "18.3.1";
        var REACT_ELEMENT_TYPE = Symbol.for("react.element");
        var REACT_PORTAL_TYPE = Symbol.for("react.portal");
        var REACT_FRAGMENT_TYPE = Symbol.for("react.fragment");
        var REACT_STRICT_MODE_TYPE = Symbol.for("react.strict_mode");
        var REACT_PROFILER_TYPE = Symbol.for("react.profiler");
        var REACT_PROVIDER_TYPE = Symbol.for("react.provider");
        var REACT_CONTEXT_TYPE = Symbol.for("react.context");
        var REACT_FORWARD_REF_TYPE = Symbol.for("react.forward_ref");
        var REACT_SUSPENSE_TYPE = Symbol.for("react.suspense");
        var REACT_SUSPENSE_LIST_TYPE = Symbol.for("react.suspense_list");
        var REACT_MEMO_TYPE = Symbol.for("react.memo");
        var REACT_LAZY_TYPE = Symbol.for("react.lazy");
        var REACT_OFFSCREEN_TYPE = Symbol.for("react.offscreen");
        var MAYBE_ITERATOR_SYMBOL = Symbol.iterator;
        var FAUX_ITERATOR_SYMBOL = "@@iterator";
        function getIteratorFn(maybeIterable) {
          if (maybeIterable === null || typeof maybeIterable !== "object") {
            return null;
          }
          var maybeIterator = MAYBE_ITERATOR_SYMBOL && maybeIterable[MAYBE_ITERATOR_SYMBOL] || maybeIterable[FAUX_ITERATOR_SYMBOL];
          if (typeof maybeIterator === "function") {
            return maybeIterator;
          }
          return null;
        }
        var ReactCurrentDispatcher = {
          /**
           * @internal
           * @type {ReactComponent}
           */
          current: null
        };
        var ReactCurrentBatchConfig = {
          transition: null
        };
        var ReactCurrentActQueue = {
          current: null,
          // Used to reproduce behavior of `batchedUpdates` in legacy mode.
          isBatchingLegacy: false,
          didScheduleLegacyUpdate: false
        };
        var ReactCurrentOwner = {
          /**
           * @internal
           * @type {ReactComponent}
           */
          current: null
        };
        var ReactDebugCurrentFrame = {};
        var currentExtraStackFrame = null;
        function setExtraStackFrame(stack) {
          {
            currentExtraStackFrame = stack;
          }
        }
        {
          ReactDebugCurrentFrame.setExtraStackFrame = function(stack) {
            {
              currentExtraStackFrame = stack;
            }
          };
          ReactDebugCurrentFrame.getCurrentStack = null;
          ReactDebugCurrentFrame.getStackAddendum = function() {
            var stack = "";
            if (currentExtraStackFrame) {
              stack += currentExtraStackFrame;
            }
            var impl = ReactDebugCurrentFrame.getCurrentStack;
            if (impl) {
              stack += impl() || "";
            }
            return stack;
          };
        }
        var enableScopeAPI = false;
        var enableCacheElement = false;
        var enableTransitionTracing = false;
        var enableLegacyHidden = false;
        var enableDebugTracing = false;
        var ReactSharedInternals = {
          ReactCurrentDispatcher,
          ReactCurrentBatchConfig,
          ReactCurrentOwner
        };
        {
          ReactSharedInternals.ReactDebugCurrentFrame = ReactDebugCurrentFrame;
          ReactSharedInternals.ReactCurrentActQueue = ReactCurrentActQueue;
        }
        function warn(format) {
          {
            {
              for (var _len = arguments.length, args = new Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {
                args[_key - 1] = arguments[_key];
              }
              printWarning("warn", format, args);
            }
          }
        }
        function error(format) {
          {
            {
              for (var _len2 = arguments.length, args = new Array(_len2 > 1 ? _len2 - 1 : 0), _key2 = 1; _key2 < _len2; _key2++) {
                args[_key2 - 1] = arguments[_key2];
              }
              printWarning("error", format, args);
            }
          }
        }
        function printWarning(level, format, args) {
          {
            var ReactDebugCurrentFrame2 = ReactSharedInternals.ReactDebugCurrentFrame;
            var stack = ReactDebugCurrentFrame2.getStackAddendum();
            if (stack !== "") {
              format += "%s";
              args = args.concat([stack]);
            }
            var argsWithFormat = args.map(function(item) {
              return String(item);
            });
            argsWithFormat.unshift("Warning: " + format);
            Function.prototype.apply.call(console[level], console, argsWithFormat);
          }
        }
        var didWarnStateUpdateForUnmountedComponent = {};
        function warnNoop(publicInstance, callerName) {
          {
            var _constructor = publicInstance.constructor;
            var componentName = _constructor && (_constructor.displayName || _constructor.name) || "ReactClass";
            var warningKey = componentName + "." + callerName;
            if (didWarnStateUpdateForUnmountedComponent[warningKey]) {
              return;
            }
            error("Can't call %s on a component that is not yet mounted. This is a no-op, but it might indicate a bug in your application. Instead, assign to `this.state` directly or define a `state = {};` class property with the desired state in the %s component.", callerName, componentName);
            didWarnStateUpdateForUnmountedComponent[warningKey] = true;
          }
        }
        var ReactNoopUpdateQueue = {
          /**
           * Checks whether or not this composite component is mounted.
           * @param {ReactClass} publicInstance The instance we want to test.
           * @return {boolean} True if mounted, false otherwise.
           * @protected
           * @final
           */
          isMounted: function(publicInstance) {
            return false;
          },
          /**
           * Forces an update. This should only be invoked when it is known with
           * certainty that we are **not** in a DOM transaction.
           *
           * You may want to call this when you know that some deeper aspect of the
           * component's state has changed but `setState` was not called.
           *
           * This will not invoke `shouldComponentUpdate`, but it will invoke
           * `componentWillUpdate` and `componentDidUpdate`.
           *
           * @param {ReactClass} publicInstance The instance that should rerender.
           * @param {?function} callback Called after component is updated.
           * @param {?string} callerName name of the calling function in the public API.
           * @internal
           */
          enqueueForceUpdate: function(publicInstance, callback, callerName) {
            warnNoop(publicInstance, "forceUpdate");
          },
          /**
           * Replaces all of the state. Always use this or `setState` to mutate state.
           * You should treat `this.state` as immutable.
           *
           * There is no guarantee that `this.state` will be immediately updated, so
           * accessing `this.state` after calling this method may return the old value.
           *
           * @param {ReactClass} publicInstance The instance that should rerender.
           * @param {object} completeState Next state.
           * @param {?function} callback Called after component is updated.
           * @param {?string} callerName name of the calling function in the public API.
           * @internal
           */
          enqueueReplaceState: function(publicInstance, completeState, callback, callerName) {
            warnNoop(publicInstance, "replaceState");
          },
          /**
           * Sets a subset of the state. This only exists because _pendingState is
           * internal. This provides a merging strategy that is not available to deep
           * properties which is confusing. TODO: Expose pendingState or don't use it
           * during the merge.
           *
           * @param {ReactClass} publicInstance The instance that should rerender.
           * @param {object} partialState Next partial state to be merged with state.
           * @param {?function} callback Called after component is updated.
           * @param {?string} Name of the calling function in the public API.
           * @internal
           */
          enqueueSetState: function(publicInstance, partialState, callback, callerName) {
            warnNoop(publicInstance, "setState");
          }
        };
        var assign = Object.assign;
        var emptyObject = {};
        {
          Object.freeze(emptyObject);
        }
        function Component(props, context, updater) {
          this.props = props;
          this.context = context;
          this.refs = emptyObject;
          this.updater = updater || ReactNoopUpdateQueue;
        }
        Component.prototype.isReactComponent = {};
        Component.prototype.setState = function(partialState, callback) {
          if (typeof partialState !== "object" && typeof partialState !== "function" && partialState != null) {
            throw new Error("setState(...): takes an object of state variables to update or a function which returns an object of state variables.");
          }
          this.updater.enqueueSetState(this, partialState, callback, "setState");
        };
        Component.prototype.forceUpdate = function(callback) {
          this.updater.enqueueForceUpdate(this, callback, "forceUpdate");
        };
        {
          var deprecatedAPIs = {
            isMounted: ["isMounted", "Instead, make sure to clean up subscriptions and pending requests in componentWillUnmount to prevent memory leaks."],
            replaceState: ["replaceState", "Refactor your code to use setState instead (see https://github.com/facebook/react/issues/3236)."]
          };
          var defineDeprecationWarning = function(methodName, info) {
            Object.defineProperty(Component.prototype, methodName, {
              get: function() {
                warn("%s(...) is deprecated in plain JavaScript React classes. %s", info[0], info[1]);
                return void 0;
              }
            });
          };
          for (var fnName in deprecatedAPIs) {
            if (deprecatedAPIs.hasOwnProperty(fnName)) {
              defineDeprecationWarning(fnName, deprecatedAPIs[fnName]);
            }
          }
        }
        function ComponentDummy() {
        }
        ComponentDummy.prototype = Component.prototype;
        function PureComponent(props, context, updater) {
          this.props = props;
          this.context = context;
          this.refs = emptyObject;
          this.updater = updater || ReactNoopUpdateQueue;
        }
        var pureComponentPrototype = PureComponent.prototype = new ComponentDummy();
        pureComponentPrototype.constructor = PureComponent;
        assign(pureComponentPrototype, Component.prototype);
        pureComponentPrototype.isPureReactComponent = true;
        function createRef() {
          var refObject = {
            current: null
          };
          {
            Object.seal(refObject);
          }
          return refObject;
        }
        var isArrayImpl = Array.isArray;
        function isArray(a) {
          return isArrayImpl(a);
        }
        function typeName(value) {
          {
            var hasToStringTag = typeof Symbol === "function" && Symbol.toStringTag;
            var type = hasToStringTag && value[Symbol.toStringTag] || value.constructor.name || "Object";
            return type;
          }
        }
        function willCoercionThrow(value) {
          {
            try {
              testStringCoercion(value);
              return false;
            } catch (e) {
              return true;
            }
          }
        }
        function testStringCoercion(value) {
          return "" + value;
        }
        function checkKeyStringCoercion(value) {
          {
            if (willCoercionThrow(value)) {
              error("The provided key is an unsupported type %s. This value must be coerced to a string before before using it here.", typeName(value));
              return testStringCoercion(value);
            }
          }
        }
        function getWrappedName(outerType, innerType, wrapperName) {
          var displayName = outerType.displayName;
          if (displayName) {
            return displayName;
          }
          var functionName = innerType.displayName || innerType.name || "";
          return functionName !== "" ? wrapperName + "(" + functionName + ")" : wrapperName;
        }
        function getContextName(type) {
          return type.displayName || "Context";
        }
        function getComponentNameFromType(type) {
          if (type == null) {
            return null;
          }
          {
            if (typeof type.tag === "number") {
              error("Received an unexpected object in getComponentNameFromType(). This is likely a bug in React. Please file an issue.");
            }
          }
          if (typeof type === "function") {
            return type.displayName || type.name || null;
          }
          if (typeof type === "string") {
            return type;
          }
          switch (type) {
            case REACT_FRAGMENT_TYPE:
              return "Fragment";
            case REACT_PORTAL_TYPE:
              return "Portal";
            case REACT_PROFILER_TYPE:
              return "Profiler";
            case REACT_STRICT_MODE_TYPE:
              return "StrictMode";
            case REACT_SUSPENSE_TYPE:
              return "Suspense";
            case REACT_SUSPENSE_LIST_TYPE:
              return "SuspenseList";
          }
          if (typeof type === "object") {
            switch (type.$$typeof) {
              case REACT_CONTEXT_TYPE:
                var context = type;
                return getContextName(context) + ".Consumer";
              case REACT_PROVIDER_TYPE:
                var provider = type;
                return getContextName(provider._context) + ".Provider";
              case REACT_FORWARD_REF_TYPE:
                return getWrappedName(type, type.render, "ForwardRef");
              case REACT_MEMO_TYPE:
                var outerName = type.displayName || null;
                if (outerName !== null) {
                  return outerName;
                }
                return getComponentNameFromType(type.type) || "Memo";
              case REACT_LAZY_TYPE: {
                var lazyComponent = type;
                var payload = lazyComponent._payload;
                var init = lazyComponent._init;
                try {
                  return getComponentNameFromType(init(payload));
                } catch (x) {
                  return null;
                }
              }
            }
          }
          return null;
        }
        var hasOwnProperty = Object.prototype.hasOwnProperty;
        var RESERVED_PROPS = {
          key: true,
          ref: true,
          __self: true,
          __source: true
        };
        var specialPropKeyWarningShown, specialPropRefWarningShown, didWarnAboutStringRefs;
        {
          didWarnAboutStringRefs = {};
        }
        function hasValidRef(config) {
          {
            if (hasOwnProperty.call(config, "ref")) {
              var getter = Object.getOwnPropertyDescriptor(config, "ref").get;
              if (getter && getter.isReactWarning) {
                return false;
              }
            }
          }
          return config.ref !== void 0;
        }
        function hasValidKey(config) {
          {
            if (hasOwnProperty.call(config, "key")) {
              var getter = Object.getOwnPropertyDescriptor(config, "key").get;
              if (getter && getter.isReactWarning) {
                return false;
              }
            }
          }
          return config.key !== void 0;
        }
        function defineKeyPropWarningGetter(props, displayName) {
          var warnAboutAccessingKey = function() {
            {
              if (!specialPropKeyWarningShown) {
                specialPropKeyWarningShown = true;
                error("%s: `key` is not a prop. Trying to access it will result in `undefined` being returned. If you need to access the same value within the child component, you should pass it as a different prop. (https://reactjs.org/link/special-props)", displayName);
              }
            }
          };
          warnAboutAccessingKey.isReactWarning = true;
          Object.defineProperty(props, "key", {
            get: warnAboutAccessingKey,
            configurable: true
          });
        }
        function defineRefPropWarningGetter(props, displayName) {
          var warnAboutAccessingRef = function() {
            {
              if (!specialPropRefWarningShown) {
                specialPropRefWarningShown = true;
                error("%s: `ref` is not a prop. Trying to access it will result in `undefined` being returned. If you need to access the same value within the child component, you should pass it as a different prop. (https://reactjs.org/link/special-props)", displayName);
              }
            }
          };
          warnAboutAccessingRef.isReactWarning = true;
          Object.defineProperty(props, "ref", {
            get: warnAboutAccessingRef,
            configurable: true
          });
        }
        function warnIfStringRefCannotBeAutoConverted(config) {
          {
            if (typeof config.ref === "string" && ReactCurrentOwner.current && config.__self && ReactCurrentOwner.current.stateNode !== config.__self) {
              var componentName = getComponentNameFromType(ReactCurrentOwner.current.type);
              if (!didWarnAboutStringRefs[componentName]) {
                error('Component "%s" contains the string ref "%s". Support for string refs will be removed in a future major release. This case cannot be automatically converted to an arrow function. We ask you to manually fix this case by using useRef() or createRef() instead. Learn more about using refs safely here: https://reactjs.org/link/strict-mode-string-ref', componentName, config.ref);
                didWarnAboutStringRefs[componentName] = true;
              }
            }
          }
        }
        var ReactElement = function(type, key, ref, self, source, owner, props) {
          var element = {
            // This tag allows us to uniquely identify this as a React Element
            $$typeof: REACT_ELEMENT_TYPE,
            // Built-in properties that belong on the element
            type,
            key,
            ref,
            props,
            // Record the component responsible for creating this element.
            _owner: owner
          };
          {
            element._store = {};
            Object.defineProperty(element._store, "validated", {
              configurable: false,
              enumerable: false,
              writable: true,
              value: false
            });
            Object.defineProperty(element, "_self", {
              configurable: false,
              enumerable: false,
              writable: false,
              value: self
            });
            Object.defineProperty(element, "_source", {
              configurable: false,
              enumerable: false,
              writable: false,
              value: source
            });
            if (Object.freeze) {
              Object.freeze(element.props);
              Object.freeze(element);
            }
          }
          return element;
        };
        function createElement(type, config, children) {
          var propName;
          var props = {};
          var key = null;
          var ref = null;
          var self = null;
          var source = null;
          if (config != null) {
            if (hasValidRef(config)) {
              ref = config.ref;
              {
                warnIfStringRefCannotBeAutoConverted(config);
              }
            }
            if (hasValidKey(config)) {
              {
                checkKeyStringCoercion(config.key);
              }
              key = "" + config.key;
            }
            self = config.__self === void 0 ? null : config.__self;
            source = config.__source === void 0 ? null : config.__source;
            for (propName in config) {
              if (hasOwnProperty.call(config, propName) && !RESERVED_PROPS.hasOwnProperty(propName)) {
                props[propName] = config[propName];
              }
            }
          }
          var childrenLength = arguments.length - 2;
          if (childrenLength === 1) {
            props.children = children;
          } else if (childrenLength > 1) {
            var childArray = Array(childrenLength);
            for (var i = 0; i < childrenLength; i++) {
              childArray[i] = arguments[i + 2];
            }
            {
              if (Object.freeze) {
                Object.freeze(childArray);
              }
            }
            props.children = childArray;
          }
          if (type && type.defaultProps) {
            var defaultProps = type.defaultProps;
            for (propName in defaultProps) {
              if (props[propName] === void 0) {
                props[propName] = defaultProps[propName];
              }
            }
          }
          {
            if (key || ref) {
              var displayName = typeof type === "function" ? type.displayName || type.name || "Unknown" : type;
              if (key) {
                defineKeyPropWarningGetter(props, displayName);
              }
              if (ref) {
                defineRefPropWarningGetter(props, displayName);
              }
            }
          }
          return ReactElement(type, key, ref, self, source, ReactCurrentOwner.current, props);
        }
        function cloneAndReplaceKey(oldElement, newKey) {
          var newElement = ReactElement(oldElement.type, newKey, oldElement.ref, oldElement._self, oldElement._source, oldElement._owner, oldElement.props);
          return newElement;
        }
        function cloneElement(element, config, children) {
          if (element === null || element === void 0) {
            throw new Error("React.cloneElement(...): The argument must be a React element, but you passed " + element + ".");
          }
          var propName;
          var props = assign({}, element.props);
          var key = element.key;
          var ref = element.ref;
          var self = element._self;
          var source = element._source;
          var owner = element._owner;
          if (config != null) {
            if (hasValidRef(config)) {
              ref = config.ref;
              owner = ReactCurrentOwner.current;
            }
            if (hasValidKey(config)) {
              {
                checkKeyStringCoercion(config.key);
              }
              key = "" + config.key;
            }
            var defaultProps;
            if (element.type && element.type.defaultProps) {
              defaultProps = element.type.defaultProps;
            }
            for (propName in config) {
              if (hasOwnProperty.call(config, propName) && !RESERVED_PROPS.hasOwnProperty(propName)) {
                if (config[propName] === void 0 && defaultProps !== void 0) {
                  props[propName] = defaultProps[propName];
                } else {
                  props[propName] = config[propName];
                }
              }
            }
          }
          var childrenLength = arguments.length - 2;
          if (childrenLength === 1) {
            props.children = children;
          } else if (childrenLength > 1) {
            var childArray = Array(childrenLength);
            for (var i = 0; i < childrenLength; i++) {
              childArray[i] = arguments[i + 2];
            }
            props.children = childArray;
          }
          return ReactElement(element.type, key, ref, self, source, owner, props);
        }
        function isValidElement(object) {
          return typeof object === "object" && object !== null && object.$$typeof === REACT_ELEMENT_TYPE;
        }
        var SEPARATOR = ".";
        var SUBSEPARATOR = ":";
        function escape(key) {
          var escapeRegex = /[=:]/g;
          var escaperLookup = {
            "=": "=0",
            ":": "=2"
          };
          var escapedString = key.replace(escapeRegex, function(match) {
            return escaperLookup[match];
          });
          return "$" + escapedString;
        }
        var didWarnAboutMaps = false;
        var userProvidedKeyEscapeRegex = /\/+/g;
        function escapeUserProvidedKey(text) {
          return text.replace(userProvidedKeyEscapeRegex, "$&/");
        }
        function getElementKey(element, index) {
          if (typeof element === "object" && element !== null && element.key != null) {
            {
              checkKeyStringCoercion(element.key);
            }
            return escape("" + element.key);
          }
          return index.toString(36);
        }
        function mapIntoArray(children, array, escapedPrefix, nameSoFar, callback) {
          var type = typeof children;
          if (type === "undefined" || type === "boolean") {
            children = null;
          }
          var invokeCallback = false;
          if (children === null) {
            invokeCallback = true;
          } else {
            switch (type) {
              case "string":
              case "number":
                invokeCallback = true;
                break;
              case "object":
                switch (children.$$typeof) {
                  case REACT_ELEMENT_TYPE:
                  case REACT_PORTAL_TYPE:
                    invokeCallback = true;
                }
            }
          }
          if (invokeCallback) {
            var _child = children;
            var mappedChild = callback(_child);
            var childKey = nameSoFar === "" ? SEPARATOR + getElementKey(_child, 0) : nameSoFar;
            if (isArray(mappedChild)) {
              var escapedChildKey = "";
              if (childKey != null) {
                escapedChildKey = escapeUserProvidedKey(childKey) + "/";
              }
              mapIntoArray(mappedChild, array, escapedChildKey, "", function(c) {
                return c;
              });
            } else if (mappedChild != null) {
              if (isValidElement(mappedChild)) {
                {
                  if (mappedChild.key && (!_child || _child.key !== mappedChild.key)) {
                    checkKeyStringCoercion(mappedChild.key);
                  }
                }
                mappedChild = cloneAndReplaceKey(
                  mappedChild,
                  // Keep both the (mapped) and old keys if they differ, just as
                  // traverseAllChildren used to do for objects as children
                  escapedPrefix + // $FlowFixMe Flow incorrectly thinks React.Portal doesn't have a key
                  (mappedChild.key && (!_child || _child.key !== mappedChild.key) ? (
                    // $FlowFixMe Flow incorrectly thinks existing element's key can be a number
                    // eslint-disable-next-line react-internal/safe-string-coercion
                    escapeUserProvidedKey("" + mappedChild.key) + "/"
                  ) : "") + childKey
                );
              }
              array.push(mappedChild);
            }
            return 1;
          }
          var child;
          var nextName;
          var subtreeCount = 0;
          var nextNamePrefix = nameSoFar === "" ? SEPARATOR : nameSoFar + SUBSEPARATOR;
          if (isArray(children)) {
            for (var i = 0; i < children.length; i++) {
              child = children[i];
              nextName = nextNamePrefix + getElementKey(child, i);
              subtreeCount += mapIntoArray(child, array, escapedPrefix, nextName, callback);
            }
          } else {
            var iteratorFn = getIteratorFn(children);
            if (typeof iteratorFn === "function") {
              var iterableChildren = children;
              {
                if (iteratorFn === iterableChildren.entries) {
                  if (!didWarnAboutMaps) {
                    warn("Using Maps as children is not supported. Use an array of keyed ReactElements instead.");
                  }
                  didWarnAboutMaps = true;
                }
              }
              var iterator = iteratorFn.call(iterableChildren);
              var step;
              var ii = 0;
              while (!(step = iterator.next()).done) {
                child = step.value;
                nextName = nextNamePrefix + getElementKey(child, ii++);
                subtreeCount += mapIntoArray(child, array, escapedPrefix, nextName, callback);
              }
            } else if (type === "object") {
              var childrenString = String(children);
              throw new Error("Objects are not valid as a React child (found: " + (childrenString === "[object Object]" ? "object with keys {" + Object.keys(children).join(", ") + "}" : childrenString) + "). If you meant to render a collection of children, use an array instead.");
            }
          }
          return subtreeCount;
        }
        function mapChildren(children, func, context) {
          if (children == null) {
            return children;
          }
          var result = [];
          var count = 0;
          mapIntoArray(children, result, "", "", function(child) {
            return func.call(context, child, count++);
          });
          return result;
        }
        function countChildren(children) {
          var n = 0;
          mapChildren(children, function() {
            n++;
          });
          return n;
        }
        function forEachChildren(children, forEachFunc, forEachContext) {
          mapChildren(children, function() {
            forEachFunc.apply(this, arguments);
          }, forEachContext);
        }
        function toArray(children) {
          return mapChildren(children, function(child) {
            return child;
          }) || [];
        }
        function onlyChild(children) {
          if (!isValidElement(children)) {
            throw new Error("React.Children.only expected to receive a single React element child.");
          }
          return children;
        }
        function createContext(defaultValue) {
          var context = {
            $$typeof: REACT_CONTEXT_TYPE,
            // As a workaround to support multiple concurrent renderers, we categorize
            // some renderers as primary and others as secondary. We only expect
            // there to be two concurrent renderers at most: React Native (primary) and
            // Fabric (secondary); React DOM (primary) and React ART (secondary).
            // Secondary renderers store their context values on separate fields.
            _currentValue: defaultValue,
            _currentValue2: defaultValue,
            // Used to track how many concurrent renderers this context currently
            // supports within in a single renderer. Such as parallel server rendering.
            _threadCount: 0,
            // These are circular
            Provider: null,
            Consumer: null,
            // Add these to use same hidden class in VM as ServerContext
            _defaultValue: null,
            _globalName: null
          };
          context.Provider = {
            $$typeof: REACT_PROVIDER_TYPE,
            _context: context
          };
          var hasWarnedAboutUsingNestedContextConsumers = false;
          var hasWarnedAboutUsingConsumerProvider = false;
          var hasWarnedAboutDisplayNameOnConsumer = false;
          {
            var Consumer = {
              $$typeof: REACT_CONTEXT_TYPE,
              _context: context
            };
            Object.defineProperties(Consumer, {
              Provider: {
                get: function() {
                  if (!hasWarnedAboutUsingConsumerProvider) {
                    hasWarnedAboutUsingConsumerProvider = true;
                    error("Rendering <Context.Consumer.Provider> is not supported and will be removed in a future major release. Did you mean to render <Context.Provider> instead?");
                  }
                  return context.Provider;
                },
                set: function(_Provider) {
                  context.Provider = _Provider;
                }
              },
              _currentValue: {
                get: function() {
                  return context._currentValue;
                },
                set: function(_currentValue) {
                  context._currentValue = _currentValue;
                }
              },
              _currentValue2: {
                get: function() {
                  return context._currentValue2;
                },
                set: function(_currentValue2) {
                  context._currentValue2 = _currentValue2;
                }
              },
              _threadCount: {
                get: function() {
                  return context._threadCount;
                },
                set: function(_threadCount) {
                  context._threadCount = _threadCount;
                }
              },
              Consumer: {
                get: function() {
                  if (!hasWarnedAboutUsingNestedContextConsumers) {
                    hasWarnedAboutUsingNestedContextConsumers = true;
                    error("Rendering <Context.Consumer.Consumer> is not supported and will be removed in a future major release. Did you mean to render <Context.Consumer> instead?");
                  }
                  return context.Consumer;
                }
              },
              displayName: {
                get: function() {
                  return context.displayName;
                },
                set: function(displayName) {
                  if (!hasWarnedAboutDisplayNameOnConsumer) {
                    warn("Setting `displayName` on Context.Consumer has no effect. You should set it directly on the context with Context.displayName = '%s'.", displayName);
                    hasWarnedAboutDisplayNameOnConsumer = true;
                  }
                }
              }
            });
            context.Consumer = Consumer;
          }
          {
            context._currentRenderer = null;
            context._currentRenderer2 = null;
          }
          return context;
        }
        var Uninitialized = -1;
        var Pending = 0;
        var Resolved = 1;
        var Rejected = 2;
        function lazyInitializer(payload) {
          if (payload._status === Uninitialized) {
            var ctor = payload._result;
            var thenable = ctor();
            thenable.then(function(moduleObject2) {
              if (payload._status === Pending || payload._status === Uninitialized) {
                var resolved = payload;
                resolved._status = Resolved;
                resolved._result = moduleObject2;
              }
            }, function(error2) {
              if (payload._status === Pending || payload._status === Uninitialized) {
                var rejected = payload;
                rejected._status = Rejected;
                rejected._result = error2;
              }
            });
            if (payload._status === Uninitialized) {
              var pending2 = payload;
              pending2._status = Pending;
              pending2._result = thenable;
            }
          }
          if (payload._status === Resolved) {
            var moduleObject = payload._result;
            {
              if (moduleObject === void 0) {
                error("lazy: Expected the result of a dynamic import() call. Instead received: %s\n\nYour code should look like: \n  const MyComponent = lazy(() => import('./MyComponent'))\n\nDid you accidentally put curly braces around the import?", moduleObject);
              }
            }
            {
              if (!("default" in moduleObject)) {
                error("lazy: Expected the result of a dynamic import() call. Instead received: %s\n\nYour code should look like: \n  const MyComponent = lazy(() => import('./MyComponent'))", moduleObject);
              }
            }
            return moduleObject.default;
          } else {
            throw payload._result;
          }
        }
        function lazy(ctor) {
          var payload = {
            // We use these fields to store the result.
            _status: Uninitialized,
            _result: ctor
          };
          var lazyType = {
            $$typeof: REACT_LAZY_TYPE,
            _payload: payload,
            _init: lazyInitializer
          };
          {
            var defaultProps;
            var propTypes;
            Object.defineProperties(lazyType, {
              defaultProps: {
                configurable: true,
                get: function() {
                  return defaultProps;
                },
                set: function(newDefaultProps) {
                  error("React.lazy(...): It is not supported to assign `defaultProps` to a lazy component import. Either specify them where the component is defined, or create a wrapping component around it.");
                  defaultProps = newDefaultProps;
                  Object.defineProperty(lazyType, "defaultProps", {
                    enumerable: true
                  });
                }
              },
              propTypes: {
                configurable: true,
                get: function() {
                  return propTypes;
                },
                set: function(newPropTypes) {
                  error("React.lazy(...): It is not supported to assign `propTypes` to a lazy component import. Either specify them where the component is defined, or create a wrapping component around it.");
                  propTypes = newPropTypes;
                  Object.defineProperty(lazyType, "propTypes", {
                    enumerable: true
                  });
                }
              }
            });
          }
          return lazyType;
        }
        function forwardRef(render) {
          {
            if (render != null && render.$$typeof === REACT_MEMO_TYPE) {
              error("forwardRef requires a render function but received a `memo` component. Instead of forwardRef(memo(...)), use memo(forwardRef(...)).");
            } else if (typeof render !== "function") {
              error("forwardRef requires a render function but was given %s.", render === null ? "null" : typeof render);
            } else {
              if (render.length !== 0 && render.length !== 2) {
                error("forwardRef render functions accept exactly two parameters: props and ref. %s", render.length === 1 ? "Did you forget to use the ref parameter?" : "Any additional parameter will be undefined.");
              }
            }
            if (render != null) {
              if (render.defaultProps != null || render.propTypes != null) {
                error("forwardRef render functions do not support propTypes or defaultProps. Did you accidentally pass a React component?");
              }
            }
          }
          var elementType = {
            $$typeof: REACT_FORWARD_REF_TYPE,
            render
          };
          {
            var ownName;
            Object.defineProperty(elementType, "displayName", {
              enumerable: false,
              configurable: true,
              get: function() {
                return ownName;
              },
              set: function(name) {
                ownName = name;
                if (!render.name && !render.displayName) {
                  render.displayName = name;
                }
              }
            });
          }
          return elementType;
        }
        var REACT_MODULE_REFERENCE;
        {
          REACT_MODULE_REFERENCE = Symbol.for("react.module.reference");
        }
        function isValidElementType(type) {
          if (typeof type === "string" || typeof type === "function") {
            return true;
          }
          if (type === REACT_FRAGMENT_TYPE || type === REACT_PROFILER_TYPE || enableDebugTracing || type === REACT_STRICT_MODE_TYPE || type === REACT_SUSPENSE_TYPE || type === REACT_SUSPENSE_LIST_TYPE || enableLegacyHidden || type === REACT_OFFSCREEN_TYPE || enableScopeAPI || enableCacheElement || enableTransitionTracing) {
            return true;
          }
          if (typeof type === "object" && type !== null) {
            if (type.$$typeof === REACT_LAZY_TYPE || type.$$typeof === REACT_MEMO_TYPE || type.$$typeof === REACT_PROVIDER_TYPE || type.$$typeof === REACT_CONTEXT_TYPE || type.$$typeof === REACT_FORWARD_REF_TYPE || // This needs to include all possible module reference object
            // types supported by any Flight configuration anywhere since
            // we don't know which Flight build this will end up being used
            // with.
            type.$$typeof === REACT_MODULE_REFERENCE || type.getModuleId !== void 0) {
              return true;
            }
          }
          return false;
        }
        function memo(type, compare) {
          {
            if (!isValidElementType(type)) {
              error("memo: The first argument must be a component. Instead received: %s", type === null ? "null" : typeof type);
            }
          }
          var elementType = {
            $$typeof: REACT_MEMO_TYPE,
            type,
            compare: compare === void 0 ? null : compare
          };
          {
            var ownName;
            Object.defineProperty(elementType, "displayName", {
              enumerable: false,
              configurable: true,
              get: function() {
                return ownName;
              },
              set: function(name) {
                ownName = name;
                if (!type.name && !type.displayName) {
                  type.displayName = name;
                }
              }
            });
          }
          return elementType;
        }
        function resolveDispatcher() {
          var dispatcher = ReactCurrentDispatcher.current;
          {
            if (dispatcher === null) {
              error("Invalid hook call. Hooks can only be called inside of the body of a function component. This could happen for one of the following reasons:\n1. You might have mismatching versions of React and the renderer (such as React DOM)\n2. You might be breaking the Rules of Hooks\n3. You might have more than one copy of React in the same app\nSee https://reactjs.org/link/invalid-hook-call for tips about how to debug and fix this problem.");
            }
          }
          return dispatcher;
        }
        function useContext(Context) {
          var dispatcher = resolveDispatcher();
          {
            if (Context._context !== void 0) {
              var realContext = Context._context;
              if (realContext.Consumer === Context) {
                error("Calling useContext(Context.Consumer) is not supported, may cause bugs, and will be removed in a future major release. Did you mean to call useContext(Context) instead?");
              } else if (realContext.Provider === Context) {
                error("Calling useContext(Context.Provider) is not supported. Did you mean to call useContext(Context) instead?");
              }
            }
          }
          return dispatcher.useContext(Context);
        }
        function useState(initialState) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useState(initialState);
        }
        function useReducer(reducer, initialArg, init) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useReducer(reducer, initialArg, init);
        }
        function useRef(initialValue) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useRef(initialValue);
        }
        function useEffect(create2, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useEffect(create2, deps);
        }
        function useInsertionEffect(create2, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useInsertionEffect(create2, deps);
        }
        function useLayoutEffect(create2, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useLayoutEffect(create2, deps);
        }
        function useCallback(callback, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useCallback(callback, deps);
        }
        function useMemo(create2, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useMemo(create2, deps);
        }
        function useImperativeHandle(ref, create2, deps) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useImperativeHandle(ref, create2, deps);
        }
        function useDebugValue(value, formatterFn) {
          {
            var dispatcher = resolveDispatcher();
            return dispatcher.useDebugValue(value, formatterFn);
          }
        }
        function useTransition() {
          var dispatcher = resolveDispatcher();
          return dispatcher.useTransition();
        }
        function useDeferredValue(value) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useDeferredValue(value);
        }
        function useId() {
          var dispatcher = resolveDispatcher();
          return dispatcher.useId();
        }
        function useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot) {
          var dispatcher = resolveDispatcher();
          return dispatcher.useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
        }
        var disabledDepth = 0;
        var prevLog;
        var prevInfo;
        var prevWarn;
        var prevError;
        var prevGroup;
        var prevGroupCollapsed;
        var prevGroupEnd;
        function disabledLog() {
        }
        disabledLog.__reactDisabledLog = true;
        function disableLogs() {
          {
            if (disabledDepth === 0) {
              prevLog = console.log;
              prevInfo = console.info;
              prevWarn = console.warn;
              prevError = console.error;
              prevGroup = console.group;
              prevGroupCollapsed = console.groupCollapsed;
              prevGroupEnd = console.groupEnd;
              var props = {
                configurable: true,
                enumerable: true,
                value: disabledLog,
                writable: true
              };
              Object.defineProperties(console, {
                info: props,
                log: props,
                warn: props,
                error: props,
                group: props,
                groupCollapsed: props,
                groupEnd: props
              });
            }
            disabledDepth++;
          }
        }
        function reenableLogs() {
          {
            disabledDepth--;
            if (disabledDepth === 0) {
              var props = {
                configurable: true,
                enumerable: true,
                writable: true
              };
              Object.defineProperties(console, {
                log: assign({}, props, {
                  value: prevLog
                }),
                info: assign({}, props, {
                  value: prevInfo
                }),
                warn: assign({}, props, {
                  value: prevWarn
                }),
                error: assign({}, props, {
                  value: prevError
                }),
                group: assign({}, props, {
                  value: prevGroup
                }),
                groupCollapsed: assign({}, props, {
                  value: prevGroupCollapsed
                }),
                groupEnd: assign({}, props, {
                  value: prevGroupEnd
                })
              });
            }
            if (disabledDepth < 0) {
              error("disabledDepth fell below zero. This is a bug in React. Please file an issue.");
            }
          }
        }
        var ReactCurrentDispatcher$1 = ReactSharedInternals.ReactCurrentDispatcher;
        var prefix;
        function describeBuiltInComponentFrame(name, source, ownerFn) {
          {
            if (prefix === void 0) {
              try {
                throw Error();
              } catch (x) {
                var match = x.stack.trim().match(/\n( *(at )?)/);
                prefix = match && match[1] || "";
              }
            }
            return "\n" + prefix + name;
          }
        }
        var reentry = false;
        var componentFrameCache;
        {
          var PossiblyWeakMap = typeof WeakMap === "function" ? WeakMap : Map;
          componentFrameCache = new PossiblyWeakMap();
        }
        function describeNativeComponentFrame(fn, construct) {
          if (!fn || reentry) {
            return "";
          }
          {
            var frame = componentFrameCache.get(fn);
            if (frame !== void 0) {
              return frame;
            }
          }
          var control;
          reentry = true;
          var previousPrepareStackTrace = Error.prepareStackTrace;
          Error.prepareStackTrace = void 0;
          var previousDispatcher;
          {
            previousDispatcher = ReactCurrentDispatcher$1.current;
            ReactCurrentDispatcher$1.current = null;
            disableLogs();
          }
          try {
            if (construct) {
              var Fake = function() {
                throw Error();
              };
              Object.defineProperty(Fake.prototype, "props", {
                set: function() {
                  throw Error();
                }
              });
              if (typeof Reflect === "object" && Reflect.construct) {
                try {
                  Reflect.construct(Fake, []);
                } catch (x) {
                  control = x;
                }
                Reflect.construct(fn, [], Fake);
              } else {
                try {
                  Fake.call();
                } catch (x) {
                  control = x;
                }
                fn.call(Fake.prototype);
              }
            } else {
              try {
                throw Error();
              } catch (x) {
                control = x;
              }
              fn();
            }
          } catch (sample) {
            if (sample && control && typeof sample.stack === "string") {
              var sampleLines = sample.stack.split("\n");
              var controlLines = control.stack.split("\n");
              var s = sampleLines.length - 1;
              var c = controlLines.length - 1;
              while (s >= 1 && c >= 0 && sampleLines[s] !== controlLines[c]) {
                c--;
              }
              for (; s >= 1 && c >= 0; s--, c--) {
                if (sampleLines[s] !== controlLines[c]) {
                  if (s !== 1 || c !== 1) {
                    do {
                      s--;
                      c--;
                      if (c < 0 || sampleLines[s] !== controlLines[c]) {
                        var _frame = "\n" + sampleLines[s].replace(" at new ", " at ");
                        if (fn.displayName && _frame.includes("<anonymous>")) {
                          _frame = _frame.replace("<anonymous>", fn.displayName);
                        }
                        {
                          if (typeof fn === "function") {
                            componentFrameCache.set(fn, _frame);
                          }
                        }
                        return _frame;
                      }
                    } while (s >= 1 && c >= 0);
                  }
                  break;
                }
              }
            }
          } finally {
            reentry = false;
            {
              ReactCurrentDispatcher$1.current = previousDispatcher;
              reenableLogs();
            }
            Error.prepareStackTrace = previousPrepareStackTrace;
          }
          var name = fn ? fn.displayName || fn.name : "";
          var syntheticFrame = name ? describeBuiltInComponentFrame(name) : "";
          {
            if (typeof fn === "function") {
              componentFrameCache.set(fn, syntheticFrame);
            }
          }
          return syntheticFrame;
        }
        function describeFunctionComponentFrame(fn, source, ownerFn) {
          {
            return describeNativeComponentFrame(fn, false);
          }
        }
        function shouldConstruct(Component2) {
          var prototype = Component2.prototype;
          return !!(prototype && prototype.isReactComponent);
        }
        function describeUnknownElementTypeFrameInDEV(type, source, ownerFn) {
          if (type == null) {
            return "";
          }
          if (typeof type === "function") {
            {
              return describeNativeComponentFrame(type, shouldConstruct(type));
            }
          }
          if (typeof type === "string") {
            return describeBuiltInComponentFrame(type);
          }
          switch (type) {
            case REACT_SUSPENSE_TYPE:
              return describeBuiltInComponentFrame("Suspense");
            case REACT_SUSPENSE_LIST_TYPE:
              return describeBuiltInComponentFrame("SuspenseList");
          }
          if (typeof type === "object") {
            switch (type.$$typeof) {
              case REACT_FORWARD_REF_TYPE:
                return describeFunctionComponentFrame(type.render);
              case REACT_MEMO_TYPE:
                return describeUnknownElementTypeFrameInDEV(type.type, source, ownerFn);
              case REACT_LAZY_TYPE: {
                var lazyComponent = type;
                var payload = lazyComponent._payload;
                var init = lazyComponent._init;
                try {
                  return describeUnknownElementTypeFrameInDEV(init(payload), source, ownerFn);
                } catch (x) {
                }
              }
            }
          }
          return "";
        }
        var loggedTypeFailures = {};
        var ReactDebugCurrentFrame$1 = ReactSharedInternals.ReactDebugCurrentFrame;
        function setCurrentlyValidatingElement(element) {
          {
            if (element) {
              var owner = element._owner;
              var stack = describeUnknownElementTypeFrameInDEV(element.type, element._source, owner ? owner.type : null);
              ReactDebugCurrentFrame$1.setExtraStackFrame(stack);
            } else {
              ReactDebugCurrentFrame$1.setExtraStackFrame(null);
            }
          }
        }
        function checkPropTypes(typeSpecs, values, location2, componentName, element) {
          {
            var has = Function.call.bind(hasOwnProperty);
            for (var typeSpecName in typeSpecs) {
              if (has(typeSpecs, typeSpecName)) {
                var error$1 = void 0;
                try {
                  if (typeof typeSpecs[typeSpecName] !== "function") {
                    var err = Error((componentName || "React class") + ": " + location2 + " type `" + typeSpecName + "` is invalid; it must be a function, usually from the `prop-types` package, but received `" + typeof typeSpecs[typeSpecName] + "`.This often happens because of typos such as `PropTypes.function` instead of `PropTypes.func`.");
                    err.name = "Invariant Violation";
                    throw err;
                  }
                  error$1 = typeSpecs[typeSpecName](values, typeSpecName, componentName, location2, null, "SECRET_DO_NOT_PASS_THIS_OR_YOU_WILL_BE_FIRED");
                } catch (ex) {
                  error$1 = ex;
                }
                if (error$1 && !(error$1 instanceof Error)) {
                  setCurrentlyValidatingElement(element);
                  error("%s: type specification of %s `%s` is invalid; the type checker function must return `null` or an `Error` but returned a %s. You may have forgotten to pass an argument to the type checker creator (arrayOf, instanceOf, objectOf, oneOf, oneOfType, and shape all require an argument).", componentName || "React class", location2, typeSpecName, typeof error$1);
                  setCurrentlyValidatingElement(null);
                }
                if (error$1 instanceof Error && !(error$1.message in loggedTypeFailures)) {
                  loggedTypeFailures[error$1.message] = true;
                  setCurrentlyValidatingElement(element);
                  error("Failed %s type: %s", location2, error$1.message);
                  setCurrentlyValidatingElement(null);
                }
              }
            }
          }
        }
        function setCurrentlyValidatingElement$1(element) {
          {
            if (element) {
              var owner = element._owner;
              var stack = describeUnknownElementTypeFrameInDEV(element.type, element._source, owner ? owner.type : null);
              setExtraStackFrame(stack);
            } else {
              setExtraStackFrame(null);
            }
          }
        }
        var propTypesMisspellWarningShown;
        {
          propTypesMisspellWarningShown = false;
        }
        function getDeclarationErrorAddendum() {
          if (ReactCurrentOwner.current) {
            var name = getComponentNameFromType(ReactCurrentOwner.current.type);
            if (name) {
              return "\n\nCheck the render method of `" + name + "`.";
            }
          }
          return "";
        }
        function getSourceInfoErrorAddendum(source) {
          if (source !== void 0) {
            var fileName = source.fileName.replace(/^.*[\\\/]/, "");
            var lineNumber = source.lineNumber;
            return "\n\nCheck your code at " + fileName + ":" + lineNumber + ".";
          }
          return "";
        }
        function getSourceInfoErrorAddendumForProps(elementProps) {
          if (elementProps !== null && elementProps !== void 0) {
            return getSourceInfoErrorAddendum(elementProps.__source);
          }
          return "";
        }
        var ownerHasKeyUseWarning = {};
        function getCurrentComponentErrorInfo(parentType) {
          var info = getDeclarationErrorAddendum();
          if (!info) {
            var parentName = typeof parentType === "string" ? parentType : parentType.displayName || parentType.name;
            if (parentName) {
              info = "\n\nCheck the top-level render call using <" + parentName + ">.";
            }
          }
          return info;
        }
        function validateExplicitKey(element, parentType) {
          if (!element._store || element._store.validated || element.key != null) {
            return;
          }
          element._store.validated = true;
          var currentComponentErrorInfo = getCurrentComponentErrorInfo(parentType);
          if (ownerHasKeyUseWarning[currentComponentErrorInfo]) {
            return;
          }
          ownerHasKeyUseWarning[currentComponentErrorInfo] = true;
          var childOwner = "";
          if (element && element._owner && element._owner !== ReactCurrentOwner.current) {
            childOwner = " It was passed a child from " + getComponentNameFromType(element._owner.type) + ".";
          }
          {
            setCurrentlyValidatingElement$1(element);
            error('Each child in a list should have a unique "key" prop.%s%s See https://reactjs.org/link/warning-keys for more information.', currentComponentErrorInfo, childOwner);
            setCurrentlyValidatingElement$1(null);
          }
        }
        function validateChildKeys(node, parentType) {
          if (typeof node !== "object") {
            return;
          }
          if (isArray(node)) {
            for (var i = 0; i < node.length; i++) {
              var child = node[i];
              if (isValidElement(child)) {
                validateExplicitKey(child, parentType);
              }
            }
          } else if (isValidElement(node)) {
            if (node._store) {
              node._store.validated = true;
            }
          } else if (node) {
            var iteratorFn = getIteratorFn(node);
            if (typeof iteratorFn === "function") {
              if (iteratorFn !== node.entries) {
                var iterator = iteratorFn.call(node);
                var step;
                while (!(step = iterator.next()).done) {
                  if (isValidElement(step.value)) {
                    validateExplicitKey(step.value, parentType);
                  }
                }
              }
            }
          }
        }
        function validatePropTypes(element) {
          {
            var type = element.type;
            if (type === null || type === void 0 || typeof type === "string") {
              return;
            }
            var propTypes;
            if (typeof type === "function") {
              propTypes = type.propTypes;
            } else if (typeof type === "object" && (type.$$typeof === REACT_FORWARD_REF_TYPE || // Note: Memo only checks outer props here.
            // Inner props are checked in the reconciler.
            type.$$typeof === REACT_MEMO_TYPE)) {
              propTypes = type.propTypes;
            } else {
              return;
            }
            if (propTypes) {
              var name = getComponentNameFromType(type);
              checkPropTypes(propTypes, element.props, "prop", name, element);
            } else if (type.PropTypes !== void 0 && !propTypesMisspellWarningShown) {
              propTypesMisspellWarningShown = true;
              var _name = getComponentNameFromType(type);
              error("Component %s declared `PropTypes` instead of `propTypes`. Did you misspell the property assignment?", _name || "Unknown");
            }
            if (typeof type.getDefaultProps === "function" && !type.getDefaultProps.isReactClassApproved) {
              error("getDefaultProps is only used on classic React.createClass definitions. Use a static property named `defaultProps` instead.");
            }
          }
        }
        function validateFragmentProps(fragment) {
          {
            var keys = Object.keys(fragment.props);
            for (var i = 0; i < keys.length; i++) {
              var key = keys[i];
              if (key !== "children" && key !== "key") {
                setCurrentlyValidatingElement$1(fragment);
                error("Invalid prop `%s` supplied to `React.Fragment`. React.Fragment can only have `key` and `children` props.", key);
                setCurrentlyValidatingElement$1(null);
                break;
              }
            }
            if (fragment.ref !== null) {
              setCurrentlyValidatingElement$1(fragment);
              error("Invalid attribute `ref` supplied to `React.Fragment`.");
              setCurrentlyValidatingElement$1(null);
            }
          }
        }
        function createElementWithValidation(type, props, children) {
          var validType = isValidElementType(type);
          if (!validType) {
            var info = "";
            if (type === void 0 || typeof type === "object" && type !== null && Object.keys(type).length === 0) {
              info += " You likely forgot to export your component from the file it's defined in, or you might have mixed up default and named imports.";
            }
            var sourceInfo = getSourceInfoErrorAddendumForProps(props);
            if (sourceInfo) {
              info += sourceInfo;
            } else {
              info += getDeclarationErrorAddendum();
            }
            var typeString;
            if (type === null) {
              typeString = "null";
            } else if (isArray(type)) {
              typeString = "array";
            } else if (type !== void 0 && type.$$typeof === REACT_ELEMENT_TYPE) {
              typeString = "<" + (getComponentNameFromType(type.type) || "Unknown") + " />";
              info = " Did you accidentally export a JSX literal instead of a component?";
            } else {
              typeString = typeof type;
            }
            {
              error("React.createElement: type is invalid -- expected a string (for built-in components) or a class/function (for composite components) but got: %s.%s", typeString, info);
            }
          }
          var element = createElement.apply(this, arguments);
          if (element == null) {
            return element;
          }
          if (validType) {
            for (var i = 2; i < arguments.length; i++) {
              validateChildKeys(arguments[i], type);
            }
          }
          if (type === REACT_FRAGMENT_TYPE) {
            validateFragmentProps(element);
          } else {
            validatePropTypes(element);
          }
          return element;
        }
        var didWarnAboutDeprecatedCreateFactory = false;
        function createFactoryWithValidation(type) {
          var validatedFactory = createElementWithValidation.bind(null, type);
          validatedFactory.type = type;
          {
            if (!didWarnAboutDeprecatedCreateFactory) {
              didWarnAboutDeprecatedCreateFactory = true;
              warn("React.createFactory() is deprecated and will be removed in a future major release. Consider using JSX or use React.createElement() directly instead.");
            }
            Object.defineProperty(validatedFactory, "type", {
              enumerable: false,
              get: function() {
                warn("Factory.type is deprecated. Access the class directly before passing it to createFactory.");
                Object.defineProperty(this, "type", {
                  value: type
                });
                return type;
              }
            });
          }
          return validatedFactory;
        }
        function cloneElementWithValidation(element, props, children) {
          var newElement = cloneElement.apply(this, arguments);
          for (var i = 2; i < arguments.length; i++) {
            validateChildKeys(arguments[i], newElement.type);
          }
          validatePropTypes(newElement);
          return newElement;
        }
        function startTransition(scope, options) {
          var prevTransition = ReactCurrentBatchConfig.transition;
          ReactCurrentBatchConfig.transition = {};
          var currentTransition = ReactCurrentBatchConfig.transition;
          {
            ReactCurrentBatchConfig.transition._updatedFibers = /* @__PURE__ */ new Set();
          }
          try {
            scope();
          } finally {
            ReactCurrentBatchConfig.transition = prevTransition;
            {
              if (prevTransition === null && currentTransition._updatedFibers) {
                var updatedFibersCount = currentTransition._updatedFibers.size;
                if (updatedFibersCount > 10) {
                  warn("Detected a large number of updates inside startTransition. If this is due to a subscription please re-write it to use React provided hooks. Otherwise concurrent mode guarantees are off the table.");
                }
                currentTransition._updatedFibers.clear();
              }
            }
          }
        }
        var didWarnAboutMessageChannel = false;
        var enqueueTaskImpl = null;
        function enqueueTask(task) {
          if (enqueueTaskImpl === null) {
            try {
              var requireString = ("require" + Math.random()).slice(0, 7);
              var nodeRequire = module && module[requireString];
              enqueueTaskImpl = nodeRequire.call(module, "timers").setImmediate;
            } catch (_err) {
              enqueueTaskImpl = function(callback) {
                {
                  if (didWarnAboutMessageChannel === false) {
                    didWarnAboutMessageChannel = true;
                    if (typeof MessageChannel === "undefined") {
                      error("This browser does not have a MessageChannel implementation, so enqueuing tasks via await act(async () => ...) will fail. Please file an issue at https://github.com/facebook/react/issues if you encounter this warning.");
                    }
                  }
                }
                var channel = new MessageChannel();
                channel.port1.onmessage = callback;
                channel.port2.postMessage(void 0);
              };
            }
          }
          return enqueueTaskImpl(task);
        }
        var actScopeDepth = 0;
        var didWarnNoAwaitAct = false;
        function act(callback) {
          {
            var prevActScopeDepth = actScopeDepth;
            actScopeDepth++;
            if (ReactCurrentActQueue.current === null) {
              ReactCurrentActQueue.current = [];
            }
            var prevIsBatchingLegacy = ReactCurrentActQueue.isBatchingLegacy;
            var result;
            try {
              ReactCurrentActQueue.isBatchingLegacy = true;
              result = callback();
              if (!prevIsBatchingLegacy && ReactCurrentActQueue.didScheduleLegacyUpdate) {
                var queue = ReactCurrentActQueue.current;
                if (queue !== null) {
                  ReactCurrentActQueue.didScheduleLegacyUpdate = false;
                  flushActQueue(queue);
                }
              }
            } catch (error2) {
              popActScope(prevActScopeDepth);
              throw error2;
            } finally {
              ReactCurrentActQueue.isBatchingLegacy = prevIsBatchingLegacy;
            }
            if (result !== null && typeof result === "object" && typeof result.then === "function") {
              var thenableResult = result;
              var wasAwaited = false;
              var thenable = {
                then: function(resolve, reject) {
                  wasAwaited = true;
                  thenableResult.then(function(returnValue2) {
                    popActScope(prevActScopeDepth);
                    if (actScopeDepth === 0) {
                      recursivelyFlushAsyncActWork(returnValue2, resolve, reject);
                    } else {
                      resolve(returnValue2);
                    }
                  }, function(error2) {
                    popActScope(prevActScopeDepth);
                    reject(error2);
                  });
                }
              };
              {
                if (!didWarnNoAwaitAct && typeof Promise !== "undefined") {
                  Promise.resolve().then(function() {
                  }).then(function() {
                    if (!wasAwaited) {
                      didWarnNoAwaitAct = true;
                      error("You called act(async () => ...) without await. This could lead to unexpected testing behaviour, interleaving multiple act calls and mixing their scopes. You should - await act(async () => ...);");
                    }
                  });
                }
              }
              return thenable;
            } else {
              var returnValue = result;
              popActScope(prevActScopeDepth);
              if (actScopeDepth === 0) {
                var _queue = ReactCurrentActQueue.current;
                if (_queue !== null) {
                  flushActQueue(_queue);
                  ReactCurrentActQueue.current = null;
                }
                var _thenable = {
                  then: function(resolve, reject) {
                    if (ReactCurrentActQueue.current === null) {
                      ReactCurrentActQueue.current = [];
                      recursivelyFlushAsyncActWork(returnValue, resolve, reject);
                    } else {
                      resolve(returnValue);
                    }
                  }
                };
                return _thenable;
              } else {
                var _thenable2 = {
                  then: function(resolve, reject) {
                    resolve(returnValue);
                  }
                };
                return _thenable2;
              }
            }
          }
        }
        function popActScope(prevActScopeDepth) {
          {
            if (prevActScopeDepth !== actScopeDepth - 1) {
              error("You seem to have overlapping act() calls, this is not supported. Be sure to await previous act() calls before making a new one. ");
            }
            actScopeDepth = prevActScopeDepth;
          }
        }
        function recursivelyFlushAsyncActWork(returnValue, resolve, reject) {
          {
            var queue = ReactCurrentActQueue.current;
            if (queue !== null) {
              try {
                flushActQueue(queue);
                enqueueTask(function() {
                  if (queue.length === 0) {
                    ReactCurrentActQueue.current = null;
                    resolve(returnValue);
                  } else {
                    recursivelyFlushAsyncActWork(returnValue, resolve, reject);
                  }
                });
              } catch (error2) {
                reject(error2);
              }
            } else {
              resolve(returnValue);
            }
          }
        }
        var isFlushing = false;
        function flushActQueue(queue) {
          {
            if (!isFlushing) {
              isFlushing = true;
              var i = 0;
              try {
                for (; i < queue.length; i++) {
                  var callback = queue[i];
                  do {
                    callback = callback(true);
                  } while (callback !== null);
                }
                queue.length = 0;
              } catch (error2) {
                queue = queue.slice(i + 1);
                throw error2;
              } finally {
                isFlushing = false;
              }
            }
          }
        }
        var createElement$1 = createElementWithValidation;
        var cloneElement$1 = cloneElementWithValidation;
        var createFactory = createFactoryWithValidation;
        var Children = {
          map: mapChildren,
          forEach: forEachChildren,
          count: countChildren,
          toArray,
          only: onlyChild
        };
        exports.Children = Children;
        exports.Component = Component;
        exports.Fragment = REACT_FRAGMENT_TYPE;
        exports.Profiler = REACT_PROFILER_TYPE;
        exports.PureComponent = PureComponent;
        exports.StrictMode = REACT_STRICT_MODE_TYPE;
        exports.Suspense = REACT_SUSPENSE_TYPE;
        exports.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED = ReactSharedInternals;
        exports.act = act;
        exports.cloneElement = cloneElement$1;
        exports.createContext = createContext;
        exports.createElement = createElement$1;
        exports.createFactory = createFactory;
        exports.createRef = createRef;
        exports.forwardRef = forwardRef;
        exports.isValidElement = isValidElement;
        exports.lazy = lazy;
        exports.memo = memo;
        exports.startTransition = startTransition;
        exports.unstable_act = act;
        exports.useCallback = useCallback;
        exports.useContext = useContext;
        exports.useDebugValue = useDebugValue;
        exports.useDeferredValue = useDeferredValue;
        exports.useEffect = useEffect;
        exports.useId = useId;
        exports.useImperativeHandle = useImperativeHandle;
        exports.useInsertionEffect = useInsertionEffect;
        exports.useLayoutEffect = useLayoutEffect;
        exports.useMemo = useMemo;
        exports.useReducer = useReducer;
        exports.useRef = useRef;
        exports.useState = useState;
        exports.useSyncExternalStore = useSyncExternalStore;
        exports.useTransition = useTransition;
        exports.version = ReactVersion;
        if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ !== "undefined" && typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.registerInternalModuleStop === "function") {
          __REACT_DEVTOOLS_GLOBAL_HOOK__.registerInternalModuleStop(new Error());
        }
      })();
    }
  }
});

// ../node_modules/react/index.js
var require_react = __commonJS({
  "../node_modules/react/index.js"(exports, module) {
    "use strict";
    if (process.env.NODE_ENV === "production") {
      module.exports = require_react_production_min();
    } else {
      module.exports = require_react_development();
    }
  }
});

// ../node_modules/zustand/esm/react.mjs
function useStore(api2, selector = identity) {
  const slice = import_react.default.useSyncExternalStore(
    api2.subscribe,
    import_react.default.useCallback(() => selector(api2.getState()), [api2, selector]),
    import_react.default.useCallback(() => selector(api2.getInitialState()), [api2, selector])
  );
  import_react.default.useDebugValue(slice);
  return slice;
}
var import_react, identity, createImpl, create;
var init_react = __esm({
  "../node_modules/zustand/esm/react.mjs"() {
    import_react = __toESM(require_react(), 1);
    init_vanilla();
    identity = (arg) => arg;
    createImpl = (createState) => {
      const api2 = createStore(createState);
      const useBoundStore = (selector) => useStore(api2, selector);
      Object.assign(useBoundStore, api2);
      return useBoundStore;
    };
    create = ((createState) => createState ? createImpl(createState) : createImpl);
  }
});

// ../node_modules/zustand/esm/index.mjs
var init_esm = __esm({
  "../node_modules/zustand/esm/index.mjs"() {
    init_vanilla();
    init_react();
  }
});

// src/api.ts
async function req(url, opts) {
  const r = await fetch(url, {
    ...opts,
    headers: { "Content-Type": "application/json", ...opts?.headers || {} }
  });
  const data = await r.json().catch(() => ({}));
  if (!r.ok || data.ok === false) throw new Error(data.error || `HTTP ${r.status}`);
  return data;
}
var API_BASE, api;
var init_api = __esm({
  "src/api.ts"() {
    "use strict";
    API_BASE = "/debug/api";
    api = {
      status: () => req(`${API_BASE}/status`),
      // 环境清单:读共享端点。合并前这里调 /api/settings(旧 debug 节的 sshs),
      // 现在 hosts 节是唯一数据源,调试页与字典页看到的是同一份环境。
      settings: () => req("/api/hosts"),
      // 保存:只发要改的节(见 HostsPatch),后端把 400 的中文 error 原样回给界面
      saveSettings: (cfg) => req("/api/hosts", { method: "PUT", body: JSON.stringify(cfg) }),
      // 自动获取数据库连接要素(SSH 上服务器探测「从服务器获取」;note 说明未获取到的原因)
      probeDB: (body) => req("/api/dbprobe", { method: "POST", body: JSON.stringify(body) }),
      // 客户端直连测试(按表单显式字段连库,凭据取账号列表首项)。
      // 共享端点要求 {connection: …} 包一层(与字典页同形),不再是平铺连接字段。
      connTest: (connection) => req("/api/conntest", { method: "POST", body: JSON.stringify({ connection }) }),
      // 账号清单「验证」:SSH 上服务器以该账号+密码连显式目标库 select 1(只读)
      dbAccVerify: (body) => req("/api/dbaccverify", { method: "POST", body: JSON.stringify(body) }),
      list: () => req(`${API_BASE}/sessions`),
      launch: (module, prog, opts) => req(`${API_BASE}/sessions`, { method: "POST", body: JSON.stringify({ module, prog, ...opts }) }),
      snapshot: (id) => req(`${API_BASE}/sessions/${id}`),
      // 结束调试:只结束本轮运行,宿主会话保留(idle),可直接再次启动
      quit: (id) => req(`${API_BASE}/sessions/${id}`, { method: "DELETE" }),
      // 会话管理(单一常驻会话)
      sessionRestart: (id) => req(`${API_BASE}/sessions/${id}/restart`, { method: "POST" }),
      sessionClose: (id) => req(`${API_BASE}/sessions/${id}/close`, { method: "POST" }),
      sessionSwitch: (env) => req(`${API_BASE}/sessions/switch`, { method: "POST", body: JSON.stringify({ env }) }),
      // 空闲态重新设置会话 TOPENT(空值 = 清除覆盖并重新下发配置默认)
      topent: (id, value) => req(`${API_BASE}/sessions/${id}/topent`, { method: "POST", body: JSON.stringify({ value }) }),
      bpAdd: (id, location2) => req(`${API_BASE}/sessions/${id}/breakpoints`, { method: "POST", body: JSON.stringify({ location: location2 }) }),
      bpDel: (id, num) => req(`${API_BASE}/sessions/${id}/breakpoints/${num}`, { method: "DELETE" }),
      control: (id, action, arg) => req(`${API_BASE}/sessions/${id}/control`, { method: "POST", body: JSON.stringify({ action, arg }) }),
      // 切模式(纯人工/协作)。人可任意方向切;AI 不能自行解除纯人工模式(服务端 403)
      setMode: (id, mode) => req(`${API_BASE}/sessions/${id}/mode`, { method: "POST", body: JSON.stringify({ mode }) }),
      print: (id, expr) => req(`${API_BASE}/sessions/${id}/print`, { method: "POST", body: JSON.stringify({ expr }) }),
      where: (id) => req(`${API_BASE}/sessions/${id}/where`, { method: "POST" }),
      raw: (id, command) => req(`${API_BASE}/sessions/${id}/raw`, { method: "POST", body: JSON.stringify({ command }) }),
      locals: (id) => req(`${API_BASE}/sessions/${id}/locals`),
      globals: (id, limit) => req(`${API_BASE}/sessions/${id}/globals${limit ? `?limit=${limit}` : ""}`),
      sources: (id) => req(`${API_BASE}/sessions/${id}/sources`),
      functions: (id, limit) => req(`${API_BASE}/sessions/${id}/functions${limit ? `?limit=${limit}` : ""}`),
      autovars: (id) => req(`${API_BASE}/sessions/${id}/autovars`),
      // 自动变量面板开关:停站后是否由服务端自动求值当前源码窗变量(默认关)
      autovarsAuto: (id, auto) => req(`${API_BASE}/sessions/${id}/autovars`, { method: "POST", body: JSON.stringify({ auto }) }),
      frame: (id, num) => req(`${API_BASE}/sessions/${id}/frame`, { method: "POST", body: JSON.stringify({ num }) }),
      bpEnabled: (id, num, enabled) => req(`${API_BASE}/sessions/${id}/breakpoints/${num}/enabled`, { method: "POST", body: JSON.stringify({ enabled }) }),
      sourceByFile: (id, file, module) => req(`${API_BASE}/sessions/${id}/source?file=${encodeURIComponent(file)}&module=${encodeURIComponent(module)}`),
      // 定位函数到源文件与行号(fgldb info line;仅停站可用)
      locate: (id, word) => req(`${API_BASE}/sessions/${id}/locate`, { method: "POST", body: JSON.stringify({ word }) }),
      // 行号校准:检测 fgldb(DVM)行号与磁盘源码的偏移(仅停站可用)
      calibrate: (id) => req(`${API_BASE}/sessions/${id}/calibrate`, { method: "POST" }),
      wsTest: (mode, url, body, soap) => req(`${API_BASE}/wstest`, { method: "POST", body: JSON.stringify({ mode, url, body, soap }) }),
      // 条件用对象传递(参数变多后位置参数太容易错位);空值不发送
      wsLogs: (q) => {
        const p4 = new URLSearchParams();
        for (const [k, v] of Object.entries(q)) {
          if (v === void 0 || v === null || v === "" || v === false) continue;
          p4.set(k, typeof v === "boolean" ? v ? "1" : "0" : String(v));
        }
        p4.set("onlyFail", q.onlyFail ? "1" : "0");
        p4.set("page", String(q.page ?? 1));
        p4.set("pageSize", "200");
        return req(`${API_BASE}/wslogs?${p4.toString()}`);
      },
      wsLogContent: (rowid) => req(`${API_BASE}/wslogs/content?rowid=${encodeURIComponent(rowid)}`),
      // 重放调试:request 非空 = 用界面里改过的入参(后端落临时文件后作为入参文件)
      wsLogDebug: (rowid, request) => req(`${API_BASE}/wslogs/debug`, {
        method: "POST",
        body: JSON.stringify(request ? { rowid, request } : { rowid })
      }),
      sourcePreview: (module, prog) => req(`${API_BASE}/source-preview?module=${encodeURIComponent(module)}&prog=${encodeURIComponent(prog)}`)
    };
  }
});

// ../shared/theme.ts
function systemPrefersDark() {
  const m = mq();
  return m ? m.matches : true;
}
function resolveDark(mode) {
  if (mode === "dark") return true;
  if (mode === "light") return false;
  return systemPrefersDark();
}
function readStoredTheme() {
  try {
    const v = localStorage.getItem(THEME_KEY);
    return v === "light" || v === "system" ? v : "dark";
  } catch {
    return "dark";
  }
}
function applyDark(dark) {
  document.documentElement.classList.toggle("dark", dark);
}
function watchSystemTheme(onChange) {
  const m = mq();
  if (!m) return;
  if (typeof m.addEventListener === "function") m.addEventListener("change", onChange);
  else if (typeof m.addListener === "function") {
    ;
    m.addListener(onChange);
  }
}
var THEME_KEY, THEME_MEDIA, mq;
var init_theme = __esm({
  "../shared/theme.ts"() {
    THEME_KEY = "tt.theme";
    THEME_MEDIA = "(prefers-color-scheme: dark)";
    mq = () => typeof window !== "undefined" && typeof window.matchMedia === "function" ? window.matchMedia(THEME_MEDIA) : null;
  }
});

// src/theme.ts
var init_theme2 = __esm({
  "src/theme.ts"() {
    "use strict";
    init_theme();
  }
});

// src/store.ts
var store_exports = {};
__export(store_exports, {
  connectWS: () => connectWS,
  useStore: () => useStore2
});
function beginRetire(id) {
  retireID = id;
  retireUntil = 0;
}
function endRetire() {
  if (retireID) retireUntil = Date.now() + RETIRE_GRACE_MS;
}
function cancelRetire() {
  retireID = null;
  retireUntil = 0;
}
function isRetiringTeardown(ev) {
  if (!retireID || ev.sessionId !== retireID) return false;
  if (retireUntil && Date.now() >= retireUntil) {
    retireID = null;
    return false;
  }
  return ev.type === "dead" || ev.type === "state" && (ev.state === "idle" || ev.state === "exit");
}
function now() {
  return (/* @__PURE__ */ new Date()).toLocaleTimeString("zh-CN", { hour12: false });
}
function scheduleReveal(set, get, file, line) {
  pendingReveal = { file, line };
  if (revealTimer) clearTimeout(revealTimer);
  revealTimer = window.setTimeout(async () => {
    revealTimer = void 0;
    const target = pendingReveal;
    pendingReveal = null;
    if (!target) return;
    if (get().activeTab !== "debug") {
      set({ loadingSource: false });
      return;
    }
    if (target.file && (target.file !== get().sourceDVM || get().sourceMissing)) {
      await get().refreshSource(target.file, target.line);
    } else if (target.line) {
      set({ loadingSource: false });
      get().reveal("debug", target.line);
    } else {
      set({ loadingSource: false });
    }
  }, REVEAL_DEBOUNCE_MS);
}
function jumpToMain(set, get) {
  if (get().currentLine > 0) return;
  const lines = get().sourceContent.split("\n");
  for (let i = 0; i < lines.length; i++) {
    if (/^\s*MAIN\b/i.test(lines[i])) {
      set({ currentLine: i + 1 });
      break;
    }
  }
}
function loadWsLogContent(item) {
  const hit = wsLogContentCache.get(item.rowid);
  if (hit) return Promise.resolve(hit);
  const flying = wsLogContentInflight.get(item.rowid);
  if (flying) return flying;
  const p4 = api.wsLogContent(item.rowid).then(({ content }) => {
    wsLogContentCache.set(item.rowid, content);
    if (wsLogContentCache.size > WSLOG_CACHE_MAX) {
      const oldest = wsLogContentCache.keys().next().value;
      if (oldest) wsLogContentCache.delete(oldest);
    }
    return content;
  }).finally(() => {
    wsLogContentInflight.delete(item.rowid);
  });
  wsLogContentInflight.set(item.rowid, p4);
  return p4;
}
function pollUntilStopped(set, get) {
  if (snapTimer) clearInterval(snapTimer);
  snapTimer = window.setInterval(async () => {
    const st = get();
    if (!st.sessionId) {
      if (snapTimer) clearInterval(snapTimer);
      return;
    }
    try {
      const snap = await api.snapshot(st.sessionId);
      const nl = !!snap.stop?.file && snap.stop.file !== st.sourceDVM;
      set({
        state: snap.state,
        stop: snap.stop,
        breakpoints: snap.breakpoints || [],
        started: !!snap.started,
        holdingSeconds: snap.holdingSeconds || 0,
        sessionEnv: snap.env || get().sessionEnv,
        module: snap.module || get().module,
        runProg: snap.runProg || get().runProg,
        currentLine: nl ? 0 : snap.stop?.line || get().currentLine,
        loadingSource: nl ? true : get().loadingSource
      });
      if (nl && snap.stop?.file) void get().refreshSource(snap.stop.file, snap.stop.line);
      else if (snap.state === "stopped" && snap.stop?.line && !nl) get().reveal("debug", snap.stop.line);
      if (snap.state === "stopped") {
        clearInterval(snapTimer);
        snapTimer = void 0;
        set({ activeTab: "debug" });
        startHoldTimer(set, get);
        if (get().stackAuto) void get().refreshFrames();
        void get().refreshWatches();
        void get().refreshSource(snap.stop?.file);
        window.setTimeout(() => {
          const cur = get();
          if (cur.sessionId && cur.state === "stopped") void cur.refreshSnapshot();
        }, 2500);
        if (snap.stop?.reason === "breakpoint") {
          st.pushTimeline({ origin: "system", kind: "stop", text: `\u547D\u4E2D\u65AD\u70B9 ${snap.stop.file}:${snap.stop.line}` });
        }
      }
      if (snap.state === "exit") {
        clearInterval(snapTimer);
        snapTimer = void 0;
      }
      if (snap.state === "idle") {
        clearInterval(snapTimer);
        snapTimer = void 0;
        stopHoldTimer();
        set({ started: false, stop: null, currentLine: 0, loadingSource: false, selectedFrame: -1, autovars: [], breakpoints: [], frames: [], adjustedBps: {} });
      }
    } catch {
      const st2 = get();
      if (!st2.sessionId) {
        if (snapTimer) {
          clearInterval(snapTimer);
          snapTimer = void 0;
        }
        return;
      }
      try {
        const { sessions } = await api.list();
        if (!sessions.some((x) => x.id === st2.sessionId)) {
          if (snapTimer) {
            clearInterval(snapTimer);
            snapTimer = void 0;
          }
          stopHoldTimer();
          set({
            sessionId: null,
            state: "",
            loadingSource: false,
            stop: null,
            launchError: st2.launchError || "\u8C03\u8BD5\u4F1A\u8BDD\u542F\u52A8\u5931\u8D25(\u4F1A\u8BDD\u5DF2\u6D88\u5931),\u8BF7\u68C0\u67E5\u4F5C\u4E1A\u540D\u6216\u670D\u52A1\u5668\u72B6\u6001"
          });
        }
      } catch {
      }
    }
  }, 1e3);
}
function alignAutoPrefs(set, get) {
  const sid = get().sessionId;
  if (!sid || !get().autovarsAuto) return;
  void api.autovarsAuto(sid, true).catch(() => {
  });
  if (get().state === "stopped") {
    void api.autovars(sid).then(({ vars }) => set({ autovars: vars || [] })).catch(() => {
    });
  }
}
function startHoldTimer(set, get) {
  stopHoldTimer();
  holdTimer = window.setInterval(() => {
    const st = get();
    if (st.state !== "stopped") {
      stopHoldTimer();
      return;
    }
    set({ holdingSeconds: st.holdingSeconds + 1 });
  }, 1e3);
}
function stopHoldTimer() {
  if (holdTimer) {
    clearInterval(holdTimer);
    holdTimer = void 0;
  }
}
function connectWS() {
  const setWs = useStore2.getState().setWsConnected;
  const onEvent = useStore2.getState().onEvent;
  const pushRaw = useStore2.getState().pushRaw;
  let closed = false;
  let ws;
  function connect() {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    ws = new WebSocket(`${proto}://${location.host}${API_BASE}/ws`);
    ws.onopen = () => {
      setWs(true);
      void useStore2.getState().syncFromSessions();
    };
    ws.onclose = () => {
      setWs(false);
      if (!closed) setTimeout(connect, 2e3);
    };
    ws.onmessage = (m) => {
      try {
        const ev = JSON.parse(m.data);
        if (ev.type === "replay" || ev.type === "hello") {
          const epoch = String(ev.epoch || "");
          if (epoch && epoch !== seenEpoch) {
            seenEpoch = epoch;
            seenSeq = 0;
          }
          const replay = useStore2.getState().onReplay;
          for (const e of ev.events || []) replay(e);
          return;
        }
        onEvent(ev);
      } catch {
        pushRaw(String(m.data));
      }
    };
  }
  connect();
  return () => {
    closed = true;
    ws?.close();
  };
}
var snapTimer, holdTimer, initialTheme, retireID, retireUntil, RETIRE_GRACE_MS, seenSeq, seenEpoch, revealTimer, pendingReveal, REVEAL_DEBOUNCE_MS, WSLOG_CACHE_MAX, wsLogContentCache, wsLogContentInflight, useStore2;
var init_store = __esm({
  "src/store.ts"() {
    "use strict";
    init_esm();
    init_api();
    init_theme2();
    initialTheme = readStoredTheme();
    retireID = null;
    retireUntil = 0;
    RETIRE_GRACE_MS = 2e3;
    seenSeq = 0;
    seenEpoch = "";
    pendingReveal = null;
    REVEAL_DEBOUNCE_MS = 150;
    WSLOG_CACHE_MAX = 30;
    wsLogContentCache = /* @__PURE__ */ new Map();
    wsLogContentInflight = /* @__PURE__ */ new Map();
    useStore2 = create((set, get) => ({
      wsConnected: false,
      sessionId: null,
      module: "",
      prog: "",
      state: "",
      sessionEnv: "",
      started: false,
      stop: null,
      holdingSeconds: 0,
      launching: false,
      launchError: "",
      showRight: true,
      showBottom: true,
      // 面板开关默认关(localStorage tt.stackAuto / tt.autovarsAuto 持久化)
      stackAuto: localStorage.getItem("tt.stackAuto") === "1",
      autovarsAuto: localStorage.getItem("tt.autovarsAuto") === "1",
      lastReplayRowid: null,
      lastReplayRequest: "",
      // 右侧边栏默认落在「会话」(先选环境再调试);Tab 顺序见 App.tsx 右侧切换栏,不做持久化
      rightView: "session",
      breakpoints: [],
      adjustedBps: {},
      frames: [],
      watches: [],
      autovars: [],
      selectedFrame: -1,
      backendDead: "",
      mode: "solo",
      inflight: null,
      timeline: [],
      rawLog: [],
      runProg: "",
      view: "debug",
      theme: initialTheme,
      dark: resolveDark(initialTheme),
      wsLogs: [],
      wsLogsLoading: false,
      wsLogsPage: 1,
      wsLogsHasMore: false,
      wsLogSel: null,
      wsLogContent: null,
      wsLogTab: "info",
      wsLogErr: "",
      wsTestMode: "3",
      wsTestUrl: "",
      wsTestBody: "",
      wsTestSoap: false,
      wsTestResult: null,
      wsTestRunning: false,
      wsTestErr: "",
      sourceContent: "",
      sourcePath: "",
      sourceDVM: "",
      sourceMissing: false,
      currentLine: 0,
      loadingSource: false,
      lineOffset: 0,
      tabs: [],
      activeTab: "debug",
      revealReq: null,
      setWsConnected: (b) => set({ wsConnected: b }),
      toggleRight: () => set((st) => ({ showRight: !st.showRight })),
      toggleBottom: () => set((st) => ({ showBottom: !st.showBottom })),
      // 调用栈面板「自动」开关(默认关):开 → 停站后自动抓调用栈,已停站时立即抓一次
      toggleStackAuto: () => {
        const st = get();
        const on = !st.stackAuto;
        localStorage.setItem("tt.stackAuto", on ? "1" : "0");
        set({ stackAuto: on });
        if (on && get().state === "stopped") void get().refreshFrames();
      },
      // 自动变量面板「自动」开关(默认关):服务端停站后自动求值窗口变量才发生(见
      // 会话 SetAutovarsOn);开 → 下发当前会话(已停站时服务端立即补一次求值并推送)
      toggleAutovarsAuto: () => {
        const st = get();
        const on = !st.autovarsAuto;
        localStorage.setItem("tt.autovarsAuto", on ? "1" : "0");
        set({ autovarsAuto: on });
        const sid = get().sessionId;
        if (!sid) return;
        void api.autovarsAuto(sid, on).catch(() => {
        });
      },
      setRightView: (v) => set({ rightView: v }),
      pushRaw: (line) => set((st) => {
        const log = st.rawLog.length > 3e3 ? st.rawLog.slice(-2e3) : st.rawLog;
        return { rawLog: [...log, line] };
      }),
      // 直接执行 fgldb 命令(复刻原版 fgldeb Ctrl+D「Input Debugger Command」的透传):
      // 命令与输出原样进协议流;状态类命令执行后刷新快照/栈,与原版行为一致
      sendRaw: async (cmd) => {
        const { sessionId } = get();
        if (!sessionId || !cmd.trim()) return;
        get().pushRaw(`> ${cmd}`);
        try {
          const r = await api.raw(sessionId, cmd);
          for (const l of r.lines) get().pushRaw(l);
          const head = cmd.trim().split(/\s+/)[0].toLowerCase();
          if (["step", "next", "continue", "run"].includes(head)) {
            pollUntilStopped(set, get);
          } else if (["break", "tbreak", "clear", "delete", "enable", "disable", "where", "frame", "finish", "return"].includes(head)) {
            void get().refreshSnapshot();
            if (get().state === "stopped") void get().refreshFrames();
          }
        } catch (e) {
          get().pushRaw(`[\u9519\u8BEF] ${e.message || String(e)}`);
        }
      },
      pushTimeline: (item) => set((st) => ({
        timeline: [...st.timeline.slice(-500), { ...item, time: now() }]
      })),
      onEvent: (ev) => {
        const st = get();
        if (st.sessionId && ev.sessionId && ev.sessionId !== st.sessionId) return;
        if (isRetiringTeardown(ev)) return;
        if (ev.seq) {
          if (ev.seq <= seenSeq) return;
          seenSeq = ev.seq;
        }
        switch (ev.type) {
          case "output":
            st.pushRaw(ev.text || "");
            return;
          case "state":
            set({ state: ev.state || "" });
            if (ev.state === "stopped") {
              if (get().stackAuto) void get().refreshFrames();
              void get().refreshWatches();
              startHoldTimer(set, get);
            }
            if (ev.state === "running" || ev.state === "idle" || ev.state === "exit") {
              if (snapTimer) {
                clearInterval(snapTimer);
                snapTimer = void 0;
              }
              stopHoldTimer();
              if (ev.state === "running") set({ selectedFrame: -1 });
            }
            if (ev.state === "idle") {
              set({ started: false, stop: null, currentLine: 0, loadingSource: false, selectedFrame: -1, autovars: [], breakpoints: [], frames: [], adjustedBps: {} });
            }
            if (ev.state === "exit") {
              st.pushTimeline({ origin: "system", kind: "warn", text: "\u4F1A\u8BDD\u5DF2\u65AD\u5F00(\u7ED3\u675F\u4F1A\u8BDD/\u5207\u6362\u73AF\u5883\u6216\u8FDE\u63A5\u4E2D\u65AD)" });
              set({ loadingSource: false, currentLine: 0, started: false, stop: null });
              void get().syncFromSessions();
            }
            return;
          case "dead":
            set({ backendDead: ev.text || "\u8C03\u8BD5\u540E\u7AEF\u8FDE\u63A5\u65AD\u5F00" });
            st.pushTimeline({ origin: "system", kind: "warn", text: ev.text || "\u8C03\u8BD5\u540E\u7AEF\u8FDE\u63A5\u65AD\u5F00" });
            return;
          case "autovars":
            set({ autovars: ev.vars || [] });
            return;
          case "stopped": {
            const nl = !!ev.stop?.file && (ev.stop.file !== get().sourceDVM || get().sourceMissing);
            const follow = get().activeTab === "debug";
            set({
              stop: ev.stop || null,
              state: "stopped",
              currentLine: nl ? 0 : ev.stop?.line || get().currentLine,
              selectedFrame: -1,
              loadingSource: follow && nl,
              ...follow ? { activeTab: "debug" } : {}
            });
            st.pushTimeline({
              origin: "system",
              kind: "stop",
              text: `\u505C\u7AD9[${ev.stop?.reason}] ${ev.stop?.file || ""}:${ev.stop?.line ?? ""} ${ev.stop?.func || ""}`
            });
            void (async () => {
              let file = ev.stop?.file;
              let frameLine = 0;
              if (!file && get().stackAuto) {
                const frames = await get().refreshFrames();
                file = frames[0]?.file;
                frameLine = frames[0]?.line ?? 0;
              }
              scheduleReveal(set, get, file, frameLine || ev.stop?.line || 0);
              await get().refreshWatches();
              await get().refreshSnapshot();
            })();
            startHoldTimer(set, get);
            return;
          }
          case "watchdog":
            st.pushTimeline({ origin: "system", kind: "warn", text: ev.text || "\u770B\u95E8\u72D7\u89E6\u53D1" });
            return;
          case "ai_action":
            st.pushTimeline({ origin: "ai", kind: "command", text: ev.text || "" });
            if (ev.action?.startsWith("bp.") || ev.action === "session.mode") void get().refreshSnapshot();
            return;
          case "log": {
            const text = ev.text || "";
            st.pushTimeline({ origin: "system", kind: "info", text });
            if (text.startsWith("\u542F\u52A8\u5931\u8D25")) {
              if (snapTimer) {
                clearInterval(snapTimer);
                snapTimer = void 0;
              }
              stopHoldTimer();
              set({ launchError: text, sessionId: null, state: "", loadingSource: false, stop: null, currentLine: 0 });
            }
            return;
          }
        }
      },
      // WS 开场补发:把历史填进时间线,让刷新页面后仍能看到"刚才发生了什么"。
      // **只追加时间线,不触发任何副作用** —— 不刷源码/栈、不起计时器;页面状态
      // 另有 refreshSnapshot 负责,两者不打架。序号去重与 onEvent 共用同一游标,
      // 所以重连时重复补发是幂等的。
      onReplay: (ev) => {
        const st = get();
        if (st.sessionId && ev.sessionId && ev.sessionId !== st.sessionId) return;
        if (ev.seq) {
          if (ev.seq <= seenSeq) return;
          seenSeq = ev.seq;
        }
        switch (ev.type) {
          case "stopped":
            st.pushTimeline({
              origin: "system",
              kind: "stop",
              text: `\u505C\u7AD9[${ev.stop?.reason}] ${ev.stop?.file || ""}:${ev.stop?.line ?? ""} ${ev.stop?.func || ""}`
            });
            break;
          case "ai_action":
            st.pushTimeline({ origin: "ai", kind: "command", text: ev.text || "" });
            break;
          case "log":
            st.pushTimeline({ origin: "system", kind: "info", text: ev.text || "" });
            break;
          case "watchdog":
            st.pushTimeline({ origin: "system", kind: "warn", text: ev.text || "\u770B\u95E8\u72D7\u89E6\u53D1" });
            break;
        }
      },
      setView: (v) => set({ view: v }),
      setLaunchError: (msg) => set({ launchError: msg }),
      // 外观:选择 → 持久化 + 立即应用(跟随系统时 dark 取当前系统偏好)
      setTheme: (t) => {
        const dark = resolveDark(t);
        localStorage.setItem(THEME_KEY, t);
        applyDark(dark);
        set({ theme: t, dark });
      },
      setActiveTab: (key) => set({ activeTab: key }),
      // 定位信号(唯一驱动视口滚动)。
      // nav=true = 纯浏览跳转(断点列表点击):只滚视口,不动 currentLine(黄色停站高亮仍留在
      // 程序真正停的那行),也不受"仅停站可定位"的调试门限制——与大纲点击同一语义。
      reveal: (key, line, nav = false) => {
        if (!(line > 0)) return;
        set((st) => ({
          revealReq: { key, line, nav, seq: (st.revealReq?.seq ?? 0) + 1 },
          currentLine: !nav && key === "debug" ? line : st.currentLine
        }));
      },
      locate: async (word) => {
        const { sessionId } = get();
        if (!sessionId) throw new Error("\u65E0\u8C03\u8BD5\u4F1A\u8BDD");
        return api.locate(sessionId, word);
      },
      // 行号校准:检测 fgldb(DVM)行号与磁盘源码的偏移,前插 offset 行让 Monaco 行号对齐协议流。
      // 用户在协议流与源码对不上时主动触发(不同文件/编译版本的偏移可能不同)
      calibrate: async () => {
        const { sessionId, state } = get();
        if (!sessionId) return;
        if (state !== "stopped") {
          get().pushTimeline({ origin: "system", kind: "warn", text: "\u884C\u53F7\u6821\u51C6:\u9700\u505C\u7AD9\u540E\u624D\u80FD\u68C0\u6D4B\u504F\u79FB" });
          return;
        }
        try {
          const { offset } = await api.calibrate(sessionId);
          if (offset < 0) {
            get().pushTimeline({ origin: "system", kind: "warn", text: `\u884C\u53F7\u6821\u51C6:\u6E90\u7801\u6BD4\u7F16\u8BD1\u7248\u672C\u591A ${-offset} \u884C,\u65E0\u6CD5\u524D\u63D2\u5BF9\u9F50,\u8BF7\u5728\u670D\u52A1\u5668\u91CD\u65B0\u7F16\u8BD1\u6216\u6838\u5BF9\u6E90\u7801\u7248\u672C` });
            return;
          }
          set({ lineOffset: offset });
          get().pushTimeline({ origin: "human", kind: "info", text: offset > 0 ? `\u884C\u53F7\u6821\u51C6\u5B8C\u6210:\u534F\u8BAE\u884C\u53F7\u6BD4\u6E90\u7801\u591A ${offset} \u884C,\u5DF2\u524D\u63D2\u5BF9\u9F50` : "\u884C\u53F7\u6821\u51C6\u5B8C\u6210:\u884C\u53F7\u65E0\u504F\u79FB" });
          const st = get();
          if (st.stop?.line) get().reveal("debug", st.stop.line);
        } catch (e) {
          get().pushTimeline({ origin: "system", kind: "warn", text: `\u884C\u53F7\u6821\u51C6\u5931\u8D25: ${e.message || String(e)}` });
        }
      },
      closeTab: (key) => set((st) => {
        const idx = st.tabs.findIndex((t) => t.key === key);
        if (idx < 0) return {};
        const tabs = st.tabs.filter((t) => t.key !== key);
        if (st.activeTab !== key) return { tabs };
        return { tabs, activeTab: tabs[idx] ? tabs[idx].key : tabs[idx - 1] ? tabs[idx - 1].key : "debug" };
      }),
      openSourceTab: async (file, line) => {
        const key = "src:" + file;
        const st = get();
        const existing = st.tabs.find((t) => t.key === key);
        if (existing) {
          set({ activeTab: key, tabs: line ? st.tabs.map((t) => t.key === key ? { ...t, line } : t) : st.tabs });
          if (line) get().reveal(key, line);
          return;
        }
        if (!st.sessionId) return;
        const tab = { key, file, content: "", path: "", line, loading: true };
        set({ tabs: [...st.tabs, tab], activeTab: key });
        try {
          const { source } = await api.sourceByFile(st.sessionId, file, st.module);
          set((s2) => ({ tabs: s2.tabs.map((t) => t.key === key ? { ...t, content: source.content, path: source.path, loading: false } : t) }));
          if (line) get().reveal(key, line);
        } catch {
          set((s2) => ({ tabs: s2.tabs.map((t) => t.key === key ? { ...t, loading: false, missing: true } : t) }));
        }
      },
      setWsTest: (p4) => {
        const patch = {};
        if (p4.mode !== void 0) patch.wsTestMode = p4.mode;
        if (p4.url !== void 0) patch.wsTestUrl = p4.url;
        if (p4.body !== void 0) patch.wsTestBody = p4.body;
        if (p4.soap !== void 0) patch.wsTestSoap = p4.soap;
        if (p4.result !== void 0) patch.wsTestResult = p4.result;
        set(patch);
      },
      runWsTest: async () => {
        const st = get();
        set({ wsTestRunning: true, wsTestErr: "" });
        try {
          const { result } = await api.wsTest(st.wsTestMode, st.wsTestUrl, st.wsTestBody, st.wsTestSoap);
          set({ wsTestResult: result });
        } catch (e) {
          set({ wsTestErr: e.message || String(e) });
        } finally {
          set({ wsTestRunning: false });
        }
      },
      loadWsLogs: async (q) => {
        set({ wsLogsLoading: true, wsLogErr: "" });
        try {
          const r = await api.wsLogs(q);
          set({ wsLogs: r.items || [], wsLogsHasMore: !!r.hasMore, wsLogsPage: q.page ?? 1, wsLogSel: null, wsLogContent: null });
        } catch (e) {
          set({ wsLogErr: e.message || String(e), wsLogs: [] });
        } finally {
          set({ wsLogsLoading: false });
        }
      },
      // 点击行即取报文(接口一次返回 request+response,不是切页签才取)。
      // 命中预取缓存时瞬时显示;同一行的并发请求会被合并(见 loadWsLogContent)。
      selectWsLog: async (item) => {
        const cached = wsLogContentCache.get(item.rowid);
        set({ wsLogSel: item, wsLogContent: cached ?? null, wsLogErr: "" });
        if (cached) return;
        try {
          const content = await loadWsLogContent(item);
          if (get().wsLogSel?.rowid === item.rowid) set({ wsLogContent: content });
        } catch (e) {
          if (get().wsLogSel?.rowid === item.rowid) set({ wsLogErr: e.message || String(e) });
        }
      },
      // 悬停预取:鼠标在行上停留一小会儿就开始取报文,等真点下去时通常已经就绪。
      // 已缓存/在途则直接返回,不产生重复请求。
      prefetchWsLog: (item) => {
        if (wsLogContentCache.has(item.rowid) || wsLogContentInflight.has(item.rowid)) return;
        void loadWsLogContent(item).then((content) => {
          if (get().wsLogSel?.rowid === item.rowid) set({ wsLogContent: content });
        }).catch(() => {
        });
      },
      setWsLogTab: (t) => set({ wsLogTab: t }),
      // 关闭日志详情面板:回到"只显示列表"的默认态
      closeWsLogDetail: () => set({ wsLogSel: null, wsLogContent: null }),
      replayDebug: async (item, request) => {
        await get().replayStart(item.rowid, request);
      },
      // 接口日志重放启动:立即切到 debug 页进 loading,再请求后端重放该日志
      // (后端按 rowid 重读报文并落临时文件,报文参数在 ArgsOverride 里,随会话保留);
      // request 非空 = 用界面改过的入参(后端写临时文件后优先使用)
      replayStart: async (rowid, request) => {
        const old = get().sessionId;
        beginRetire(old);
        set({
          view: "debug",
          wsLogErr: "",
          launchError: "",
          launching: true,
          state: "loading",
          sessionId: null,
          timeline: [],
          rawLog: [],
          watches: [],
          autovars: [],
          backendDead: "",
          selectedFrame: -1,
          stop: null,
          frames: [],
          breakpoints: [],
          sourceContent: "",
          sourcePath: "",
          sourceDVM: "",
          sourceMissing: false,
          currentLine: 0,
          lineOffset: 0,
          loadingSource: true,
          lastReplayRowid: rowid,
          lastReplayRequest: request ?? ""
        });
        try {
          if (old) void api.quit(old).catch(() => {
          });
          const r = await api.wsLogDebug(rowid, request);
          const mod = r.module || "";
          const rp = r.runProg || r.prog || "";
          set({
            sessionId: r.sessionId,
            module: mod,
            prog: r.prog || rp,
            runProg: rp,
            state: "loading",
            sourceDVM: "",
            sourceMissing: false,
            currentLine: 0,
            loadingSource: true
          });
          pollUntilStopped(set, get);
        } catch (e) {
          cancelRetire();
          set({ wsLogErr: e.message || String(e), state: "", launchError: `\u542F\u52A8\u5931\u8D25: ${e.message}`, sessionId: null, loadingSource: false, stop: null });
        } finally {
          endRetire();
          set({ launching: false });
        }
      },
      launch: async (module, prog) => {
        set({ launching: true, launchError: "", timeline: [], rawLog: [], watches: [], autovars: [], backendDead: "", selectedFrame: -1, lastReplayRowid: null, lastReplayRequest: "" });
        set({ sourceContent: "", sourcePath: "", sourceDVM: "", sourceMissing: false, currentLine: 0, loadingSource: true });
        beginRetire(get().sessionId);
        try {
          const r = await api.launch(module, prog);
          const mod = r.module || module;
          const rp = r.runProg || prog;
          set({ sessionId: r.sessionId, module: mod, prog, runProg: rp, state: "loading" });
          get().pushTimeline({ origin: "human", kind: "command", text: `\u542F\u52A8\u8C03\u8BD5\u4F1A\u8BDD ${mod}/${prog}` });
          alignAutoPrefs(set, get);
          pollUntilStopped(set, get);
        } catch (e) {
          cancelRetire();
          get().pushTimeline({ origin: "system", kind: "warn", text: `\u542F\u52A8\u5931\u8D25: ${e.message}` });
          set({ launchError: `\u542F\u52A8\u5931\u8D25: ${e.message}`, sessionId: null, state: "", loadingSource: false, stop: null, currentLine: 0 });
          throw e;
        } finally {
          endRetire();
          set({ launching: false });
        }
      },
      refreshSnapshot: async () => {
        const { sessionId } = get();
        if (!sessionId) return;
        try {
          const snap = await api.snapshot(sessionId);
          set({
            state: snap.state,
            stop: snap.stop,
            breakpoints: snap.breakpoints || [],
            started: !!snap.started,
            holdingSeconds: snap.holdingSeconds || 0,
            sessionEnv: snap.env || get().sessionEnv,
            // 谁在驾驶 + 正在执行哪条命令(界面按它收敛写操作、显示 inflight)
            mode: snap.mode === "collab" ? "collab" : "solo",
            inflight: snap.inflight || null,
            // 免模块启动时后端会按作业名解析模块,回读给前端(源码兜底路径依赖它)
            module: snap.module || get().module,
            runProg: snap.runProg || get().runProg
          });
          if (snap.state === "stopped" && !snapTimer) startHoldTimer(set, get);
          if (snap.state !== "stopped" && snap.state !== "loading" && snapTimer) {
            clearInterval(snapTimer);
            snapTimer = void 0;
          }
        } catch {
        }
      },
      refreshSource: async (file, line) => {
        const { sessionId, module, prog, runProg, sourceDVM } = get();
        if (!sessionId) return;
        let f = file || get().stop?.file;
        let entryMode = false;
        if (!f) {
          const rp = runProg || prog;
          if (!rp) return;
          f = `${module ? module + "_" : ""}${rp}.4gl`;
          entryMode = true;
        } else if (get().stop?.reason === "entry") {
          entryMode = true;
        }
        if (f === sourceDVM && !get().sourceMissing) {
          if (line) set({ currentLine: line });
          return;
        }
        set({ loadingSource: true, currentLine: 0, lineOffset: 0 });
        try {
          const { source } = await api.sourceByFile(sessionId, f, module);
          set({ sourceContent: source.content, sourcePath: source.path, sourceDVM: f, sourceMissing: false });
          const st = get();
          if (line) {
            set({ currentLine: line });
            if (!entryMode) get().reveal("debug", line);
          } else if (entryMode && st.currentLine === 0) {
            jumpToMain(set, get);
            st.pushTimeline({ origin: "system", kind: "info", text: "\u5165\u53E3\u505C\u7AD9:\u5DF2\u663E\u793A\u6E90\u7801,\u70B9\u51FB\u884C\u53F7\u4E0B\u65AD\u70B9\u540E\u70B9\u300C\u7EE7\u7EED F5\u300D\u5F00\u59CB" });
          }
          if (entryMode && get().currentLine > 0) get().reveal("debug", get().currentLine);
        } catch {
          if (entryMode) {
            try {
              const { source } = await api.sourceByFile(sessionId, `c${module}_${runProg || prog}.4gl`, module);
              set({ sourceContent: source.content, sourcePath: source.path, sourceDVM: f, sourceMissing: false });
              if (get().currentLine === 0) {
                jumpToMain(set, get);
                get().pushTimeline({ origin: "system", kind: "info", text: "\u5165\u53E3\u505C\u7AD9:\u5DF2\u663E\u793A\u5BA2\u5236\u6E90\u7801,\u70B9\u51FB\u884C\u53F7\u4E0B\u65AD\u70B9\u540E\u70B9\u300C\u7EE7\u7EED F5\u300D\u5F00\u59CB" });
              }
              if (get().currentLine > 0) get().reveal("debug", get().currentLine);
              return;
            } catch {
            }
          }
          set({ sourceContent: "", sourcePath: "", sourceDVM: f, sourceMissing: true });
          if (line) set({ currentLine: line });
        } finally {
          set({ loadingSource: false });
        }
      },
      refreshFrames: async () => {
        const { sessionId, state } = get();
        if (!sessionId || state !== "stopped") return [];
        try {
          const { frames } = await api.where(sessionId);
          set({ frames });
          return frames;
        } catch {
          set({ frames: [] });
          return [];
        }
      },
      refreshWatches: async () => {
        const { sessionId, watches, state } = get();
        if (!sessionId || state !== "stopped") return;
        const out = [];
        for (const w of watches) {
          try {
            const { value } = await api.print(sessionId, w.expr);
            out.push({ expr: w.expr, value });
          } catch (e) {
            out.push({ expr: w.expr, error: e.message || String(e) });
          }
        }
        set({ watches: out });
      },
      addWatch: async (expr) => {
        const st = get();
        if (!expr.trim() || st.watches.some((w) => w.expr === expr)) return;
        set({ watches: [...st.watches, { expr }] });
        await st.refreshWatches();
        st.pushTimeline({ origin: "human", kind: "command", text: `\u76D1\u89C6 ${expr}` });
      },
      removeWatch: (expr) => set((st) => ({ watches: st.watches.filter((w) => w.expr !== expr) })),
      control: async (action, arg) => {
        const { sessionId } = get();
        if (!sessionId) return;
        get().pushTimeline({ origin: "human", kind: "command", text: arg ? `${action} ${arg}` : action });
        try {
          const resp = await api.control(sessionId, action, arg);
          if (action !== "interrupt") {
            await get().refreshSnapshot();
            const st = get();
            if (st.state === "stopped" && get().stackAuto) void st.refreshFrames();
          }
          if (action === "run" || action === "continue") pollUntilStopped(set, get);
          void resp;
        } catch (e) {
          get().pushTimeline({ origin: "system", kind: "warn", text: `${action} \u5931\u8D25: ${e.message}` });
        }
      },
      // 切模式(纯人工/协作)。人可任意方向切 —— 这是"协作模式下不被锁死"的逃生舱口;
      // AI 侧不能自行解除纯人工模式(服务端会 403)。
      setMode: async (mode) => {
        const { sessionId } = get();
        if (!sessionId) return;
        try {
          await api.setMode(sessionId, mode);
          await get().refreshSnapshot();
          get().pushTimeline({
            origin: "human",
            kind: "command",
            text: mode === "collab" ? "\u4EA4\u7ED9 AI:\u534F\u4F5C\u6A21\u5F0F" : "\u63A5\u7BA1:\u7EAF\u4EBA\u5DE5\u6A21\u5F0F"
          });
        } catch (e) {
          get().pushTimeline({ origin: "system", kind: "warn", text: `\u5207\u6362\u6A21\u5F0F\u5931\u8D25: ${e.message}` });
        }
      },
      toggleBreakpoint: async (line) => {
        const st = get();
        if (!st.sessionId) return;
        const tabFile = st.activeTab !== "debug" ? st.tabs.find((t) => t.key === st.activeTab)?.file : "";
        const inTab = (f) => !!f && !!tabFile && f.replace(/^.*[\/]/, "").toLowerCase() === tabFile.toLowerCase();
        const existing = tabFile ? st.breakpoints.find((b) => inTab(b.file) && b.line === line) : st.breakpoints.find((b) => b.line === line) || (st.adjustedBps[line] !== void 0 ? st.breakpoints.find((b) => b.num === st.adjustedBps[line]) : void 0);
        try {
          let newBp;
          if (existing) {
            await api.bpDel(st.sessionId, existing.num);
            st.pushTimeline({ origin: "human", kind: "command", text: `\u5220\u9664\u65AD\u70B9 ${existing.file}:${existing.line}` });
          } else {
            const loc = tabFile ? `${tabFile}:${line}` : st.stop?.file ? `${st.stop.file}:${line}` : String(line);
            const { breakpoint } = await api.bpAdd(st.sessionId, loc);
            newBp = breakpoint;
            st.pushTimeline({ origin: "human", kind: "command", text: `\u65AD\u70B9 ${breakpoint.file}:${breakpoint.line}` });
          }
          await st.refreshSnapshot();
          const next = {};
          for (const [clickLine, num] of Object.entries(st.adjustedBps)) {
            if (get().breakpoints.some((b) => b.num === num)) next[Number(clickLine)] = num;
          }
          if (newBp && newBp.line !== line) next[line] = newBp.num;
          set({ adjustedBps: next });
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `\u65AD\u70B9\u64CD\u4F5C\u5931\u8D25: ${e.message}` });
        }
      },
      removeBreakpoint: async (num) => {
        const st = get();
        if (!st.sessionId) return;
        try {
          await api.bpDel(st.sessionId, num);
          await st.refreshSnapshot();
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `\u5220\u9664\u65AD\u70B9\u5931\u8D25: ${e.message}` });
        }
      },
      // 选择栈帧:切换 print/locals 求值上下文并跳转该帧源码位置(仅停站时可用)
      selectFrame: async (idx) => {
        const st = get();
        if (!st.sessionId || st.state !== "stopped") return;
        const frame = st.frames.find((f) => f.idx === idx);
        try {
          await api.frame(st.sessionId, idx);
          set({ selectedFrame: idx });
          if (frame) {
            st.pushTimeline({ origin: "human", kind: "command", text: `\u9009\u5E27 #${idx} ${frame.func} ${frame.file}:${frame.line}` });
            if (frame.file && frame.file !== st.sourceDVM) await st.refreshSource(frame.file, frame.line);
            else if (frame.line) get().reveal("debug", frame.line);
          }
          void st.refreshWatches();
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `\u9009\u5E27\u5931\u8D25: ${e.message}` });
        }
      },
      toggleBPEnabled: async (num, enabled) => {
        const st = get();
        if (!st.sessionId) return;
        try {
          await api.bpEnabled(st.sessionId, num, enabled);
          st.pushTimeline({ origin: "human", kind: "command", text: `${enabled ? "\u542F\u7528" : "\u7981\u7528"}\u65AD\u70B9 #${num}` });
          await st.refreshSnapshot();
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `\u65AD\u70B9\u542F\u505C\u5931\u8D25: ${e.message}` });
        }
      },
      // 断点列表点击跳转:必要时切换到断点所在文件,再定位到断点行。
      // 纯浏览跳转(nav):只滚视口,不把黄色停站高亮拽到断点行(不改变运行上下文)
      jumpToBp: async (b) => {
        const st = get();
        if (b.file && b.file !== st.sourceDVM) await st.refreshSource(b.file);
        get().reveal("debug", b.line, true);
      },
      // 运行到光标:当前文件即停站文件时用行号,否则带文件名(fgldb until [file:]line)
      runToCursor: async (line) => {
        const st = get();
        if (!st.sessionId || st.state !== "stopped") return;
        const cur = st.sourceDVM;
        const top = st.stop?.file || st.frames[0]?.file || "";
        const arg = cur && top && cur !== top ? `${cur}:${line}` : String(line);
        await st.control("until", arg);
      },
      quit: async () => {
        const st = get();
        if (!st.sessionId) return;
        let kept = true;
        try {
          const r = await api.quit(st.sessionId);
          kept = r?.kept !== false;
          st.pushTimeline({ origin: "human", kind: "command", text: "\u7ED3\u675F\u8C03\u8BD5(\u4F1A\u8BDD\u4FDD\u7559,\u53EF\u76F4\u63A5\u518D\u6B21\u542F\u52A8)" });
        } catch {
        }
        if (snapTimer) {
          clearInterval(snapTimer);
          snapTimer = void 0;
        }
        stopHoldTimer();
        const base = {
          started: false,
          stop: null,
          breakpoints: [],
          frames: [],
          adjustedBps: {},
          currentLine: 0,
          autovars: [],
          selectedFrame: -1
        };
        if (kept) {
          set({ state: "idle", ...base });
        } else {
          set({
            sessionId: null,
            module: "",
            prog: "",
            runProg: "",
            sessionEnv: "",
            state: "",
            ...base,
            backendDead: "",
            tabs: [],
            activeTab: "debug",
            sourceContent: "",
            sourcePath: "",
            sourceDVM: "",
            sourceMissing: false,
            lineOffset: 0
          });
        }
      },
      restart: async () => {
        const st = get();
        if (!st.sessionId) return;
        const { module, prog, lastReplayRowid, lastReplayRequest } = st;
        st.pushTimeline({ origin: "human", kind: "command", text: `\u91CD\u65B0\u5F00\u59CB ${module}/${prog}${lastReplayRowid ? lastReplayRequest ? "(\u6309\u63A5\u53E3\u65E5\u5FD7\u91CD\u653E,\u7528\u6539\u8FC7\u7684\u5165\u53C2)" : "(\u6309\u539F\u63A5\u53E3\u65E5\u5FD7\u91CD\u653E)" : ""}` });
        try {
          await api.quit(st.sessionId);
        } catch {
        }
        if (snapTimer) {
          clearInterval(snapTimer);
          snapTimer = void 0;
        }
        stopHoldTimer();
        set({ state: "idle", started: false, stop: null, breakpoints: [], frames: [], adjustedBps: {}, autovars: [], selectedFrame: -1, currentLine: 0 });
        if (lastReplayRowid) await get().replayStart(lastReplayRowid, lastReplayRequest || void 0);
        else await get().launch(module, prog);
      },
      doPrint: async (expr) => {
        const st = get();
        if (!st.sessionId || !expr.trim()) return;
        try {
          const { value } = await api.print(st.sessionId, expr);
          st.pushTimeline({ origin: "human", kind: "info", text: `print ${expr} \u2192 ${value}` });
          await st.addWatch(expr);
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `print ${expr} \u2192 ${e.message}` });
        }
      },
      adoptExisting: async () => {
        try {
          const { sessions } = await api.list();
          if (!sessions.length) return;
          const s = sessions[0];
          set({ sessionId: s.id, module: s.module, prog: s.prog, runProg: s.runProg || s.prog, state: s.state, sessionEnv: s.env || "" });
          if (s.state === "idle") {
            get().pushTimeline({ origin: "system", kind: "info", text: `\u4F1A\u8BDD\u7A7A\u95F2(\u73AF\u5883 ${s.env || "\u9ED8\u8BA4"}),\u53EF\u76F4\u63A5\u542F\u52A8\u8C03\u8BD5` });
            return;
          }
          get().pushTimeline({ origin: "system", kind: "info", text: `\u63A5\u7BA1\u5DF2\u5B58\u5728\u7684\u4F1A\u8BDD ${s.module}/${s.prog}` });
          const snap = await api.snapshot(s.id);
          set({
            stop: snap.stop,
            breakpoints: snap.breakpoints || [],
            started: !!snap.started,
            currentLine: snap.stop?.line ?? 0,
            holdingSeconds: snap.holdingSeconds || 0
          });
          if (snap.state === "stopped") await get().refreshSource(snap.stop?.file);
          if (get().stackAuto) void get().refreshFrames();
          void get().refreshWatches();
          alignAutoPrefs(set, get);
        } catch {
        }
      },
      // 会话管理:右侧「会话」sheet 的 结束/重启/切换(后端会先结束当前 debug)
      sessionOp: async (op, env) => {
        const st = get();
        const label = op === "close" ? "\u7ED3\u675F\u4F1A\u8BDD" : op === "restart" ? "\u91CD\u542F\u4F1A\u8BDD" : "\u5207\u6362\u4F1A\u8BDD";
        try {
          if (op === "close" && st.sessionId) await api.sessionClose(st.sessionId);
          else if (op === "restart" && st.sessionId) await api.sessionRestart(st.sessionId);
          else if (op === "switch" && env) await api.sessionSwitch(env);
          st.pushTimeline({ origin: "human", kind: "command", text: label + (env ? " \u2192 " + env : "") });
        } catch (e) {
          st.pushTimeline({ origin: "system", kind: "warn", text: `${label}\u5931\u8D25: ${e.message || String(e)}` });
          throw e;
        }
        await get().syncFromSessions();
      },
      // 会话被替换/关闭后,按服务端现状对齐本地绑定(源码保留便于浏览)
      syncFromSessions: async () => {
        if (snapTimer) {
          clearInterval(snapTimer);
          snapTimer = void 0;
        }
        stopHoldTimer();
        const clearRun = {
          started: false,
          stop: null,
          currentLine: 0,
          breakpoints: [],
          adjustedBps: {},
          frames: [],
          autovars: [],
          selectedFrame: -1,
          holdingSeconds: 0
        };
        try {
          const { sessions } = await api.list();
          const s0 = sessions[0];
          if (!s0) {
            set({ sessionId: null, module: "", prog: "", runProg: "", sessionEnv: "", state: "", ...clearRun, loadingSource: false });
            return;
          }
          set({
            sessionId: s0.id,
            module: s0.module,
            prog: s0.prog,
            runProg: s0.runProg || s0.prog,
            sessionEnv: s0.env || "",
            state: s0.state
          });
          if (s0.state === "idle" || s0.state === "loading") {
            set({ ...clearRun, loadingSource: false });
            return;
          }
          if (s0.state === "stopped") {
            const snap = await api.snapshot(s0.id);
            set({ stop: snap.stop, breakpoints: snap.breakpoints || [], started: !!snap.started, holdingSeconds: snap.holdingSeconds || 0 });
            if (snap.stop) void get().refreshSource(snap.stop.file, snap.stop.line);
            if (get().stackAuto) void get().refreshFrames();
            void get().refreshWatches();
            alignAutoPrefs(set, get);
          }
        } catch {
        }
      }
    }));
    watchSystemTheme(() => {
      const st = useStore2.getState();
      if (st.theme !== "system") return;
      const dark = resolveDark("system");
      applyDark(dark);
      useStore2.setState({ dark });
    });
    if (typeof window !== "undefined") {
      ;
      window.__store = useStore2;
    }
    window.__store = useStore2;
  }
});

// scripts/store.test.mjs
import assert from "node:assert/strict";
globalThis.window = { setInterval: () => 1, clearInterval: () => {
}, setTimeout: () => 1 };
globalThis.localStorage = { getItem: () => null, setItem: () => {
}, removeItem: () => {
} };
globalThis.document = { documentElement: { classList: { toggle: () => {
} } } };
var release = () => {
};
globalThis.fetch = () => new Promise((res) => {
  release = () => res({
    ok: true,
    status: 200,
    json: async () => ({ ok: true, sessionId: "sess-1", module: "awsq990", prog: "awsq990", runProg: "awsq990" })
  });
});
var { useStore: useStore3 } = await Promise.resolve().then(() => (init_store(), store_exports));
var S = () => useStore3.getState();
useStore3.setState({
  sessionId: "sess-1",
  view: "wslogs",
  loadingSource: false,
  state: "stopped",
  sourceContent: "MAIN\nEND MAIN",
  timeline: [{ t: "1", origin: "system", kind: "info", text: "\u65E7\u4F1A\u8BDD\u65E5\u5FD7" }],
  stop: { file: "old.4gl", line: 7 },
  frames: [{ n: 0, file: "old.4gl", line: 7, func: "main" }]
});
var p = S().replayStart("rowid-1");
assert.equal(S().view, "debug", "\u5E94\u7ACB\u5373\u5207\u5230 debug \u9875");
assert.equal(S().loadingSource, true, "\u5E94\u7ACB\u5373\u8FDB\u5165\u52A0\u8F7D\u6001(\u4E0D\u7B49\u540E\u7AEF\u54CD\u5E94)");
assert.equal(S().state, "loading");
assert.equal(S().sessionId, null, "\u65E7\u4F1A\u8BDD\u5F15\u7528\u5E94\u7ACB\u5373\u89E3\u9664");
assert.equal(S().sourceContent, "", "\u65E7\u6E90\u7801\u5E94\u6E05\u7A7A");
assert.equal(S().timeline.length, 0, "\u65E7\u65F6\u95F4\u7EBF\u5E94\u6E05\u7A7A");
assert.equal(S().stop, null);
assert.equal(S().launching, true);
console.log("OK 1) \u70B9\u51FB\u77AC\u95F4\u5373 loading,\u65E7\u8C03\u8BD5\u73B0\u573A\u5DF2\u6E05\u7A7A");
for (const ev of [
  { type: "state", sessionId: "sess-1", state: "idle" },
  { type: "state", sessionId: "sess-1", state: "exit" },
  { type: "dead", sessionId: "sess-1", text: "SSH \u8FDE\u63A5\u65AD\u5F00" },
  { type: "log", sessionId: "sess-1", text: "\u91CD\u653E\u8C03\u8BD5\u542F\u52A8,\u7ED3\u675F\u5F53\u524D\u8C03\u8BD5(\u4F1A\u8BDD\u4FDD\u7559)" }
]) S().onEvent(ev);
assert.equal(S().loadingSource, true, "\u5728\u9014\u6536\u5C3E\u4E8B\u4EF6\u4E0D\u5F97\u53D6\u6D88\u52A0\u8F7D\u6001");
assert.equal(S().state, "loading");
assert.equal(S().backendDead, "", "\u4E0D\u5F97\u5F39\u51FA\u300C\u540E\u7AEF\u65AD\u5F00\u300D\u8BEF\u62A5\u6A2A\u5E45");
assert.equal(S().timeline.length, 1, "\u666E\u901A\u65E5\u5FD7(\u975E\u6536\u5C3E)\u4ECD\u5E94\u8FDB\u5165\u65F6\u95F4\u7EBF");
console.log("OK 2) \u5728\u9014\u65E7\u4F1A\u8BDD\u6536\u5C3E\u4E8B\u4EF6\u88AB\u4E22\u5F03,\u52A0\u8F7D\u6001\u4E0E\u6A2A\u5E45\u4E0D\u53D7\u5F71\u54CD");
release();
await p;
assert.equal(S().sessionId, "sess-1");
assert.equal(S().loadingSource, true, "\u54CD\u5E94\u540E\u4ECD\u5E94\u4FDD\u6301\u52A0\u8F7D,\u7B49\u5165\u53E3\u505C\u7AD9");
assert.equal(S().state, "loading");
console.log("OK 3) \u54CD\u5E94\u540E\u7ED1\u5B9A\u4F1A\u8BDD\u5E76\u4FDD\u6301\u52A0\u8F7D");
S().onEvent({ type: "state", sessionId: "sess-1", state: "idle" });
S().onEvent({ type: "dead", sessionId: "sess-1", text: "SSH \u8FDE\u63A5\u65AD\u5F00" });
assert.equal(S().loadingSource, true, "\u5BBD\u9650\u671F\u5185\u8FDF\u5230\u4E8B\u4EF6\u4E0D\u5F97\u53D6\u6D88\u52A0\u8F7D\u6001");
assert.equal(S().backendDead, "");
console.log("OK 4) \u54CD\u5E94\u540E\u5BBD\u9650\u671F\u5185\u8FDF\u5230\u4E8B\u4EF6\u88AB\u4E22\u5F03");
var realNow = Date.now;
Date.now = () => realNow() + 6e4;
S().onEvent({ type: "state", sessionId: "sess-1", state: "idle" });
Date.now = realNow;
assert.equal(S().loadingSource, false, "\u5BBD\u9650\u671F\u8FC7\u540E\u771F\u5B9E idle \u5FC5\u987B\u751F\u6548");
assert.equal(S().state, "idle");
console.log("OK 5) \u5BBD\u9650\u671F\u7ED3\u675F\u540E\u771F\u5B9E\u6536\u5C3E\u4E8B\u4EF6\u6B63\u5E38\u751F\u6548");
useStore3.setState({ sessionId: null, view: "wslogs", loadingSource: false, sourceContent: "" });
var p2 = S().replayStart("rowid-2");
assert.equal(S().loadingSource, true);
assert.equal(S().view, "debug");
release();
await p2;
assert.equal(S().sessionId, "sess-1");
console.log("OK 6) \u65E0\u65E7\u4F1A\u8BDD\u65F6\u540C\u6837\u7ACB\u5373 loading");
useStore3.setState({ sessionId: null, view: "debug", loadingSource: false, lastReplayRowid: "rowid-2" });
var p3 = S().launch("awsq990", "awsq990");
assert.equal(S().loadingSource, true, "launch \u5E94\u7ACB\u5373\u8FDB\u5165\u52A0\u8F7D\u6001");
assert.equal(S().lastReplayRowid, null, "\u666E\u901A\u542F\u52A8\u5E94\u6E05\u6389\u91CD\u653E\u6807\u8BB0");
release();
await p3;
console.log("OK 7) launch \u540C\u6837\u70B9\u4E0B\u5373 loading");
useStore3.setState({
  sessionId: "sess-1",
  state: "stopped",
  sourceDVM: "awsq990.4gl",
  currentLine: 120,
  stop: { file: "awsq990.4gl", line: 120 },
  revealReq: null
});
await S().jumpToBp({ num: 1, enabled: true, file: "awsq990.4gl", line: 480, func: "main" });
assert.equal(S().currentLine, 120, "\u9EC4\u8272\u505C\u7AD9\u9AD8\u4EAE\u5FC5\u987B\u7559\u5728\u771F\u5B9E\u505C\u7AD9\u884C");
assert.equal(S().revealReq.line, 480, "\u5E94\u53D1\u51FA\u6EDA\u52A8\u5230\u65AD\u70B9\u884C\u7684\u4FE1\u53F7");
assert.equal(S().revealReq.nav, true, "\u65AD\u70B9\u8DF3\u8F6C\u5E94\u6807\u8BB0\u4E3A\u7EAF\u6D4F\u89C8\u8DF3\u8F6C");
assert.equal(S().revealReq.key, "debug");
useStore3.setState({ state: "running" });
await S().jumpToBp({ num: 1, enabled: true, file: "awsq990.4gl", line: 480, func: "main" });
assert.equal(S().currentLine, 120, "\u8FD0\u884C\u4E2D\u8DF3\u8F6C\u540C\u6837\u4E0D\u5F97\u6539\u505C\u7AD9\u9AD8\u4EAE");
assert.equal(S().revealReq.nav, true);
S().reveal("debug", 500);
assert.equal(S().currentLine, 500, "\u505C\u7AD9/\u6B65\u8FDB\u5B9A\u4F4D\u5E94\u7167\u65E7\u642C\u52A8\u9EC4\u8272\u9AD8\u4EAE");
assert.equal(S().revealReq.nav, false);
console.log("OK 8) \u65AD\u70B9\u8DF3\u8F6C\u53EA\u6EDA\u89C6\u53E3,\u9EC4\u8272\u505C\u7AD9\u9AD8\u4EAE\u4E0D\u52A8");
var srcMode = "ok";
globalThis.fetch = async (url) => {
  if (String(url).includes("/source")) {
    if (srcMode === "fail") return { ok: false, status: 500, json: async () => ({ error: "\u6682\u65F6\u53D6\u4E0D\u5230" }) };
    return {
      ok: true,
      status: 200,
      json: async () => ({ ok: true, source: { path: "/u1/asf/4gl/asf_x.4gl", content: "MAIN\nEND MAIN", dvmFile: "asf_x.4gl" } })
    };
  }
  return { ok: true, status: 200, json: async () => ({ ok: true }) };
};
useStore3.setState({
  sessionId: "sx",
  module: "asf",
  prog: "bsft001_wf",
  activeTab: "debug",
  sourceDVM: "",
  sourceMissing: false,
  sourceContent: "",
  loadingSource: false
});
srcMode = "fail";
await S().refreshSource("asf_x.4gl");
assert.equal(S().sourceDVM, "asf_x.4gl", "\u53D6\u4E0D\u5230\u65F6\u4ECD\u8981\u8BB0 DVM(\u5426\u5219\u505C\u7AD9\u884C\u4F1A\u753B\u5728\u65E7\u6587\u4EF6\u4E0A)");
assert.equal(S().sourceContent, "", "\u53D6\u4E0D\u5230\u5C31\u662F\u7A7A");
assert.equal(S().loadingSource, false, "\u5931\u8D25\u4E5F\u5FC5\u987B\u9000\u51FA\u52A0\u8F7D\u6001,\u4E0D\u80FD\u4E00\u76F4\u8F6C\u5708");
assert.equal(S().sourceMissing, true, "\u8981\u6807\u8BB0 missing,\u4F9B\u540E\u7EED\u91CD\u8BD5\u5224\u65AD");
srcMode = "ok";
await S().refreshSource("asf_x.4gl");
assert.equal(S().sourceContent, "MAIN\nEND MAIN", "\u670D\u52A1\u5668\u6062\u590D\u540E\u5FC5\u987B\u91CD\u65B0\u52A0\u8F7D(\u65E7\u5B9E\u73B0\u8FD9\u91CC\u6C38\u4E45\u7A7A\u767D)");
assert.equal(S().sourceMissing, false);
console.log("OK 9) \u6E90\u7801\u53D6\u4E0D\u5230\u540E\u6062\u590D\u80FD\u91CD\u65B0\u52A0\u8F7D(missing \u4E0D\u518D\u88AB sourceDVM \u65E9\u9000\u541E\u6389)");
var pending = [];
globalThis.window.setTimeout = (fn) => {
  pending.push(fn);
  return pending.length;
};
globalThis.window.clearTimeout = () => {
};
useStore3.setState({
  sessionId: "sx",
  activeTab: "debug",
  state: "running",
  sourceDVM: "asf_x.4gl",
  sourceMissing: false,
  sourceContent: "MAIN\nEND MAIN",
  loadingSource: false
});
S().onEvent({ type: "stopped", sessionId: "sx", stop: { reason: "breakpoint", file: "bsft001_wf.4gl", line: 12 } });
assert.equal(S().loadingSource, true, "\u8DE8\u6587\u4EF6\u505C\u7AD9\u7684\u77AC\u95F4\u5E94\u8FDB\u5165\u52A0\u8F7D\u6001");
useStore3.setState({ sourceDVM: "bsft001_wf.4gl", sourceContent: "MAIN\nEND MAIN" });
for (const fn of pending.splice(0)) await fn();
assert.equal(S().loadingSource, false, '\u53BB\u6296\u8D70"\u540C\u6587\u4EF6\u53EA\u843D\u5149\u6807"\u5206\u652F\u65F6\u4E5F\u5FC5\u987B\u6536\u6389\u52A0\u8F7D\u6001');
console.log("OK 10) \u505C\u7AD9\u843D\u4F4D\u4E0D\u518D\u628A\u52A0\u8F7D\u6001\u6C38\u4E45\u7559\u5728\u754C\u9762\u4E0A");
console.log("\n\u5168\u90E8\u901A\u8FC7: 10/10");
/*! Bundled license information:

react/cjs/react.production.min.js:
  (**
   * @license React
   * react.production.min.js
   *
   * Copyright (c) Facebook, Inc. and its affiliates.
   *
   * This source code is licensed under the MIT license found in the
   * LICENSE file in the root directory of this source tree.
   *)

react/cjs/react.development.js:
  (**
   * @license React
   * react.development.js
   *
   * Copyright (c) Facebook, Inc. and its affiliates.
   *
   * This source code is licensed under the MIT license found in the
   * LICENSE file in the root directory of this source tree.
   *)
*/
