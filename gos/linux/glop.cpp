#include <algorithm>
#include <cstdio>
#include <map>
#include <mutex>
#include <set>
#include <string>
#include <sys/time.h>
#include <unistd.h>
#include <vector>

#include <X11/Xlib.h>
#include <GL/glx.h>
#include "include/glop.h"

#define LOGGING_LEVEL_FATAL 4
#define LOGGING_LEVEL_ERROR 3
#define LOGGING_LEVEL_WARN  2
#define LOGGING_LEVEL_DEBUG 1

// By default, only DEBUG messages are suppressed
#define LOGGING_LEVEL LOGGING_LEVEL_WARN

#define DO_LOG(label, fmtstring, ...) fprintf(stderr, label ": " fmtstring __VA_OPT__(,) __VA_ARGS__)

#if LOGGING_LEVEL <= LOGGING_LEVEL_FATAL
#define LOG_FATAL(args...) DO_LOG("FATAL", args)
#else
#define LOG_FATAL(fmtstring, args...) do{}while(false)
#endif // LOGGING_LEVEL <= LOGGING_LEVEL_FATAL

#if LOGGING_LEVEL <= LOGGING_LEVEL_ERROR
#define LOG_ERROR(args...) DO_LOG("ERROR", args)
#else
#define LOG_ERROR(fmtstring, args...) do{}while(false)
#endif // LOGGING_LEVEL <= LOGGING_LEVEL_ERROR

#if LOGGING_LEVEL <= LOGGING_LEVEL_WARN
#define LOG_WARN(args...) DO_LOG("WARN", args)
#else
#define LOG_WARN(args...) do{}while(false)
#endif // LOGGING_LEVEL <= LOGGING_LEVEL_WARN

#if LOGGING_LEVEL <= LOGGING_LEVEL_DEBUG
#define LOG_DEBUG(args...) DO_LOG("DEBUG", args)
#else
#define LOG_DEBUG(args...) do{}while(false)
#endif // LOGGING_LEVEL <= LOGGING_LEVEL_DEBUG

extern "C" {

void GlopClearKeyEvent(struct GlopKeyEvent* event) {
  event->index = 0;
  event->device_type = 0;
  event->press_amt = 0;
  event->timestamp = 0;
  event->cursor_x = 0;
  event->cursor_y = 0;
  event->num_lock = 0;
  event->caps_lock = 0;
}

typedef short GlopKey;

static std::mutex initMut;
Display *display = NULL;
int screen = 0;
XIM xim = NULL;
Atom close_atom;

Display *get_x_display() { return display; }
int get_x_screen() { return screen; }

static long long gtm() {
  struct timeval tv;
  // TODO(tmckee): use clock_gettime with CLOCK_MONOTONIC instead
  gettimeofday(&tv, NULL);
  // TODO(tmckee:#25) we're overflowing the long long when multiplying; we
  // should change this entire function. Currently, this is resulting in
  // 'current time' being negative.
  return (long long)tv.tv_sec * 1000000 + tv.tv_usec;
}
static int gt() {
  return gtm() / 1000;
}

struct OsWindowData {
  OsWindowData() { window = (Window)NULL; }
  ~OsWindowData() {
    glXDestroyContext(display, context);
    XDestroyIC(inputcontext);
    XFree(vinfo);
    XDestroyWindow(display, window);
  }

  Window window;
  XVisualInfo *vinfo;
  GLXContext context;
  std::vector<struct GlopKeyEvent> events;
  XIC inputcontext;
};

unsigned long GetNativeHandle(GlopWindowHandle hdl) {
  return hdl.data->window;
}

int64_t GlopInit() {
  auto lck = std::unique_lock(initMut);
  if(display == NULL) {
    display = XOpenDisplay(NULL);
    if(display == NULL) {
      LOG_FATAL("couldn't open X display\n");
      abort();
    }

    screen = DefaultScreen(display);

    xim = XOpenIM(display, NULL, NULL, NULL);
    if(xim == NULL) {
      LOG_FATAL("couldn't open X input method\n");
      abort();
    }

    close_atom = XInternAtom(display, "WM_DELETE_WINDOW", False);
  }

  return gt();
}

void glopShutDown() {
  XCloseIM(xim);
  XCloseDisplay(display);
}

XKeyEvent const * toKeyEvent(XEvent const & evt) {
  switch(evt.type) {
    case KeyPress:
    case KeyRelease:
      return &evt.xkey;
    default:
      return NULL;
  }
}

static std::pair<int, int> XCoordToGlopCoord(XWindowAttributes const * attrs, int x, int y) {
  return std::make_pair(x, attrs->height - 1 - y);
}

static bool SynthKey(XWindowAttributes const * attrs, KeySym const &sym, bool pushed, XKeyEvent const &event, Window window, struct GlopKeyEvent *ev) {
  // TODO(tmckee): this case conversion stuff is poopy.
  KeySym throwaway_lower, key;
  XConvertCase(sym, &throwaway_lower, &key);

  GlopKey ki = 0;
  switch(key) {
    case XK_A: ki = tolower('A'); break;
    case XK_B: ki = tolower('B'); break;
    case XK_C: ki = tolower('C'); break;
    case XK_D: ki = tolower('D'); break;
    case XK_E: ki = tolower('E'); break;
    case XK_F: ki = tolower('F'); break;
    case XK_G: ki = tolower('G'); break;
    case XK_H: ki = tolower('H'); break;
    case XK_I: ki = tolower('I'); break;
    case XK_J: ki = tolower('J'); break;
    case XK_K: ki = tolower('K'); break;
    case XK_L: ki = tolower('L'); break;
    case XK_M: ki = tolower('M'); break;
    case XK_N: ki = tolower('N'); break;
    case XK_O: ki = tolower('O'); break;
    case XK_P: ki = tolower('P'); break;
    case XK_Q: ki = tolower('Q'); break;
    case XK_R: ki = tolower('R'); break;
    case XK_S: ki = tolower('S'); break;
    case XK_T: ki = tolower('T'); break;
    case XK_U: ki = tolower('U'); break;
    case XK_V: ki = tolower('V'); break;
    case XK_W: ki = tolower('W'); break;
    case XK_X: ki = tolower('X'); break;
    case XK_Y: ki = tolower('Y'); break;
    case XK_Z: ki = tolower('Z'); break;

    case XK_0: ki = '0'; break;
    case XK_1: ki = '1'; break;
    case XK_2: ki = '2'; break;
    case XK_3: ki = '3'; break;
    case XK_4: ki = '4'; break;
    case XK_5: ki = '5'; break;
    case XK_6: ki = '6'; break;
    case XK_7: ki = '7'; break;
    case XK_8: ki = '8'; break;
    case XK_9: ki = '9'; break;

    case XK_F1: ki = kKeyF1; break;
    case XK_F2: ki = kKeyF2; break;
    case XK_F3: ki = kKeyF3; break;
    case XK_F4: ki = kKeyF4; break;
    case XK_F5: ki = kKeyF5; break;
    case XK_F6: ki = kKeyF6; break;
    case XK_F7: ki = kKeyF7; break;
    case XK_F8: ki = kKeyF8; break;
    case XK_F9: ki = kKeyF9; break;
    case XK_F10: ki = kKeyF10; break;
    case XK_F11: ki = kKeyF11; break;
    case XK_F12: ki = kKeyF12; break;

    case XK_KP_0: ki = kKeyPad0; break;
    case XK_KP_1: ki = kKeyPad1; break;
    case XK_KP_2: ki = kKeyPad2; break;
    case XK_KP_3: ki = kKeyPad3; break;
    case XK_KP_4: ki = kKeyPad4; break;
    case XK_KP_5: ki = kKeyPad5; break;
    case XK_KP_6: ki = kKeyPad6; break;
    case XK_KP_7: ki = kKeyPad7; break;
    case XK_KP_8: ki = kKeyPad8; break;
    case XK_KP_9: ki = kKeyPad9; break;

    case XK_Left: ki = kKeyLeft; break;
    case XK_Right: ki = kKeyRight; break;
    case XK_Up: ki = kKeyUp; break;
    case XK_Down: ki = kKeyDown; break;

    case XK_BackSpace: ki = kKeyBackspace; break;
    case XK_Tab: ki = kKeyTab; break;
    case XK_KP_Enter: ki = kKeyPadEnter; break;
    case XK_Return: ki = kKeyReturn; break;
    case XK_Escape: ki = kKeyEscape; break;

    case XK_Shift_L: ki = kKeyLeftShift; break;
    case XK_Shift_R: ki = kKeyRightShift; break;
    case XK_Control_L: ki = kKeyLeftControl; break;
    case XK_Control_R: ki = kKeyRightControl; break;
    case XK_Alt_L: ki = kKeyLeftAlt; break;
    case XK_Alt_R: ki = kKeyRightAlt; break;
    case XK_Super_L: ki = kKeyLeftGui; break;
    case XK_Super_R: ki = kKeyRightGui; break;

    case XK_KP_Divide: ki = kKeyPadDivide; break;
    case XK_KP_Multiply: ki = kKeyPadMultiply; break;
    case XK_KP_Subtract: ki = kKeyPadSubtract; break;
    case XK_KP_Add: ki = kKeyPadAdd; break;

    case XK_dead_grave: ki = '`'; break;
    case XK_minus: ki = '-'; break;
    case XK_equal: ki = '='; break;
    case XK_bracketleft: ki = '['; break;
    case XK_bracketright: ki = ']'; break;
    case XK_backslash: ki = '\\'; break;
    case XK_semicolon: ki = ';'; break;
    case XK_dead_acute: ki = '\''; break;
    case XK_comma: ki = ','; break;
    case XK_period: ki = '.'; break;
    case XK_slash: ki = '/'; break;
    // TODO(tmckee:#10): we probably want to map space to space, not to slash,
    // right?!
    case XK_space: ki = '/'; break;
  }

  if(ki == 0)
    return false;

  ev->index = ki;
  ev->device_type = glopDeviceKeyboard;
  ev->press_amt = pushed ? 1.0 : 0.0;
  ev->timestamp = gt();

  std::tie(ev->cursor_x, ev->cursor_y) = XCoordToGlopCoord(attrs, event.x, event.y);

  ev->num_lock = event.state & (1 << 4);
  ev->caps_lock = event.state & LockMask;
  return true;
}

XButtonEvent const * toButtonEvent(XEvent const & evt) {
  switch(evt.type) {
    case ButtonPress:
    case ButtonRelease:
      return &evt.xbutton;
    default:
      return NULL;
  }
}

static bool SynthButton(XWindowAttributes const *attrs, bool pushed, XButtonEvent const &event, Window window, struct GlopKeyEvent *ev) {
  bool negate = false;

  GlopKey ki;
  switch(event.button) {
    case Button1:
      ki = kMouseLButton;
      break;
    case Button2:
      ki = kMouseMButton;
      break;
    case Button3:
      ki = kMouseRButton;
      break;
    case Button5:
      negate = true;
    case Button4:
      ki = kMouseWheelVertical;
      break;
    case 7:
      negate = true;
    case 6:
      ki = kMouseWheelHorizontal;
      break;
    default:
      LOG_DEBUG("SynthButton: unknown button: %d\n", event.button);
      return false;
  }

  ev->index = ki;
  ev->device_type = glopDeviceMouse;
  ev->press_amt = pushed ? 1.0 : 0.0;
  if(negate) {
    ev->press_amt *= -1;
  }
  ev->timestamp = gt();

  std::tie(ev->cursor_x, ev->cursor_y) = XCoordToGlopCoord(attrs, event.x, event.y);
  LOG_DEBUG("SynthButton: cx/cy: %d/%d\n", ev->cursor_x, ev->cursor_y);

  ev->num_lock = event.state & (1 << 4);
  ev->caps_lock = event.state & LockMask;
  return true;
}

static bool SynthMotion(XWindowAttributes const * attrs, const XMotionEvent &event, Window window, struct GlopKeyEvent *ev, struct GlopKeyEvent *ev2) {
  ev->index = kMouseXAxis;
  ev->device_type = glopDeviceMouse;
  ev->press_amt = event.x;
  ev->timestamp = gt();
  std::tie(ev->cursor_x, ev->cursor_y) = XCoordToGlopCoord(attrs, event.x, event.y);
  ev->num_lock = event.state & (1 << 4);
  ev->caps_lock = event.state & LockMask;

  *ev2 = *ev;
  ev2->index = kMouseYAxis;
  ev2->press_amt = event.y;

  return true;
}
Bool EventTester(Display *display, XEvent *event, XPointer arg) {
  // arg == *OsWindowData
  // select for events targeted at this window
  OsWindowData *data = (OsWindowData*)(arg);
  return event->xany.window == data->window;
}

int64_t GlopThink(GlopWindowHandle windowHandle) {
  // TODO(tmckee)(clean): rename data -> windowData or smth
  OsWindowData *data = windowHandle.data;
  XEvent event;
  int last_botched_release = -1;
  int last_botched_time = -1;

  XWindowAttributes attrs;
  Status ok = XGetWindowAttributes(display, data->window, &attrs);
  if(!ok) {
    LOG_FATAL("couldn't XGetWindowAttributes\n");
    abort();
  }

  // TODO(tmckee): would using XCheck[Typed]WindowEvent be cleaner?
  while(XCheckIfEvent(display, &event, &EventTester, XPointer(data))) {
    if((event.type == KeyPress || event.type == KeyRelease) && event.xkey.keycode < 256) {
      // X is kind of a cock and likes to send us hardware repeat messages for people holding buttons down. Why do you do this, X? Why do you have to make me hate you?

      // So here's an algorithm ripped from some other source
      char kiz[32];
      XQueryKeymap(display, kiz);
      if(kiz[event.xkey.keycode >> 3] & (1 << (event.xkey.keycode % 8))) {
        if(event.type == KeyRelease) {
          last_botched_release = event.xkey.keycode;
          last_botched_time = event.xkey.time;
          continue;
        } else {
          if(last_botched_release == int(event.xkey.keycode) && unsigned(last_botched_time) == event.xkey.time) {
            // ffffffffff
            last_botched_release = -1;
            last_botched_time = -1;
            continue;
          }
        }
      }
    }

    last_botched_release = -1;
    last_botched_time = -1;

    struct GlopKeyEvent ev;
    GlopClearKeyEvent(&ev);
    switch(event.type) {
      case KeyPress:
      case KeyRelease: {
        char buf[2];
        KeySym sym;

        XLookupString(&event.xkey, buf, sizeof(buf), &sym, NULL);

        if(SynthKey(&attrs, sym, event.type == KeyPress, event.xkey, data->window, &ev))
          data->events.push_back(ev);
        break;
      }

      case ButtonPress:
      case ButtonRelease:
        LOG_DEBUG("ButtonPress/Release: event.xbutton: %d event.type: %d\n", event.xbutton.button, event.type);
        if(SynthButton(&attrs, event.type == ButtonPress, event.xbutton, data->window, &ev))
          data->events.push_back(ev);
        break;

      case MotionNotify: {
        struct GlopKeyEvent ev2;
        GlopClearKeyEvent(&ev2);
        if(SynthMotion(&attrs, event.xmotion, data->window, &ev, &ev2)) {
          data->events.push_back(ev);
          data->events.push_back(ev2);
        }
        break;
      }
      case FocusIn:
        XSetICFocus(data->inputcontext);
        break;

      case FocusOut:
        XUnsetICFocus(data->inputcontext);
        break;

      // TODO(tmckee): need to handle MappingNotify events and call
      // XRefreshKeyboardMapping as needed. See:
      // https://tronche.com/gui/x/xlib/utilities/keyboard/XRefreshKeyboardMapping.html
      case DestroyNotify:
        // WindowDashDestroy(); // ffffff
        // LOGF("destroed\n");
        // IIUC, the application gets this _once_ a window is destroyed, so we
        // shouldn't need to do anything special here?
        LOG_WARN("GlopThink: unhandled event type (DestroyNotify)\n");
        break;

      case ClientMessage :
        // TODO(tmckee#22): ummm... should we close the window? We should
        // verify that this case runs if someone clicks the 'x' through
        // window-decoration or w/e

        // IIUC, the window manager could XSendEvent to us for any number of
        // reasons but the 'close_atom' can be used to detect a "PLEASE GO
        // AWAY" message.
        if(event.xclient.format == 32 && event.xclient.data.l[0] == static_cast<long>(close_atom)) {
          LOG_WARN("Window Manager close request received but ignored\n");
          // WindowDashDestroy();
          return gt();
        }

        LOG_WARN("GlopThink: unhandled event type (ClientMessage)\n");
        break;
    }
  }

  return gt();
}

void GlopSetTitle(OsWindowData* data, const std::string& title) {
  XStoreName(display, data->window, title.c_str());
}

void glopSetCurrentContext(OsWindowData* data) {
  if(!glXMakeCurrent(display, data->window, data->context)) {
    LOG_FATAL("glxMakeCurrent failed\n");
    exit(1);
  }
}

void showConfig(GLXFBConfig const & cfg, char * out, int out_size) {
  auto end = out + out_size;
  int attribs[31] = {
    GLX_FBCONFIG_ID,
    GLX_BUFFER_SIZE,
    GLX_LEVEL,
    GLX_DOUBLEBUFFER,
    GLX_STEREO,
    GLX_AUX_BUFFERS,
    GLX_RED_SIZE,
    GLX_GREEN_SIZE,
    GLX_BLUE_SIZE,
    GLX_ALPHA_SIZE,
    GLX_DEPTH_SIZE,
    GLX_STENCIL_SIZE,
    GLX_ACCUM_RED_SIZE,
    GLX_ACCUM_GREEN_SIZE,
    GLX_ACCUM_BLUE_SIZE,
    GLX_ACCUM_ALPHA_SIZE,
    GLX_RENDER_TYPE,
    GLX_DRAWABLE_TYPE,
    GLX_X_RENDERABLE,
    GLX_VISUAL_ID,
    GLX_X_VISUAL_TYPE,
    GLX_CONFIG_CAVEAT,
    GLX_TRANSPARENT_TYPE,
    GLX_TRANSPARENT_INDEX_VALUE,
    GLX_TRANSPARENT_RED_VALUE,
    GLX_TRANSPARENT_GREEN_VALUE,
    GLX_TRANSPARENT_BLUE_VALUE,
    GLX_TRANSPARENT_ALPHA_VALUE,
    GLX_MAX_PBUFFER_WIDTH,
    GLX_MAX_PBUFFER_HEIGHT,
    GLX_MAX_PBUFFER_PIXELS,
  };

  int val;
  *out++ = '"';
  for(int attrib : attribs) {
    glXGetFBConfigAttrib(display, cfg, attrib, &val);
    out += sprintf(out, "%d, ", val);
    if(out >= end) {
      return;
    }
  }
  *out++ = '"';
}

typedef GLXContext (*glXCreateContextAttribsARBProc)(Display*, GLXFBConfig, GLXContext, Bool, const int*);

GLXContext createContextFromConfig(GLXFBConfig *fbConfig) {
  glXCreateContextAttribsARBProc glXCreateContextAttribsARB = 0;
  glXCreateContextAttribsARB = (glXCreateContextAttribsARBProc)
    glXGetProcAddressARB((const GLubyte *) "glXCreateContextAttribsARB");

  int context_attribs[] = {
    GLX_CONTEXT_MAJOR_VERSION_ARB, 4,
    GLX_CONTEXT_MINOR_VERSION_ARB, 5,
    //GLX_CONTEXT_FLAGS_ARB        , GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
    None
  };

  GLXContext noSharedContext = 0;
  Bool useDirectRendering = True;
  GLXContext ret = glXCreateContextAttribsARB(display, *fbConfig, noSharedContext, useDirectRendering, context_attribs);
  if(ret == 0) {
    LOG_FATAL("couldn't glXCreateContextAttribsARB\n");
    abort();
  }
  return ret;
}

GLXFBConfig* pickFbConfig(int* numConfigs) {
  int fbAttrs[] = {
    GLX_DOUBLEBUFFER, True,
    GLX_RED_SIZE, 8,
    GLX_GREEN_SIZE, 8,
    GLX_BLUE_SIZE, 8,
    GLX_ALPHA_SIZE, 8,
    GLX_X_RENDERABLE, True,
    GLX_X_VISUAL_TYPE, GLX_TRUE_COLOR,
    GLX_DEPTH_SIZE, 24,
    GLX_STENCIL_SIZE, 8,
    0
  };

  GLXFBConfig *fbConfig = glXChooseFBConfig(display, screen, fbAttrs, numConfigs);
  char buf[4096] = {0};
  for(int i = 0; i < *numConfigs; ++i) {
    showConfig(fbConfig[i], buf, sizeof(buf));
    LOG_WARN("config %d: '%s'\n", i, buf);
  }

  return fbConfig;
}

GlopWindowHandle GlopCreateWindowHandle(char const* title, int x, int y, int width, int height) {
  OsWindowData *nw = new OsWindowData();

  if(x <= 0 || y <= 0 || width <= 0 || height <= 0) {
    LOG_FATAL("bad window dims: (x,y): (%d,%d), (dx,dy): (%d,%d)\n", x, y, width, height);
    abort();
  }

  int numConfigs;
  GLXFBConfig *fbConfigs = pickFbConfig(&numConfigs);
  LOG_WARN("got numConfigs %d\n", numConfigs);
  if(fbConfigs == NULL || numConfigs <= 0) {
    LOG_FATAL("couldn't choose a framebuffer config. numConfigs: %d\n", numConfigs);
    abort();
  }

  GLXContext shareList = NULL;
  nw->context = createContextFromConfig(fbConfigs);

  // Grab the VisualInfo associated with the frame buffer config we chose.
  nw->vinfo = glXGetVisualFromFBConfig(display, fbConfigs[0]);

  XFree(fbConfigs);

  // TODO(tmckee): Use GLFW for window management so that we can do something like
  // glfwSetFramebufferSizeCallback(glfwWindow, setViewportOnResize)
  // Shoutout to @Hermie02 for the suggestion!

  // Define the window attributes
  XSetWindowAttributes attribs;
  attribs.event_mask = KeyPressMask | KeyReleaseMask | ButtonPressMask |
    ButtonReleaseMask | ButtonMotionMask | PointerMotionMask | FocusChangeMask
    | FocusChangeMask | ButtonPressMask | ButtonReleaseMask | ButtonMotionMask
    | PointerMotionMask | KeyPressMask | KeyReleaseMask | StructureNotifyMask |
    EnterWindowMask | LeaveWindowMask;
  attribs.colormap = XCreateColormap( display, RootWindow(display, screen), nw->vinfo->visual, AllocNone);


  nw->window = XCreateWindow(display, RootWindow(display, screen), x, y, width, height, 0, nw->vinfo->depth, InputOutput, nw->vinfo->visual, CWColormap | CWEventMask, &attribs); // I don't know if I need anything further here

  {
    Atom WMHintsAtom = XInternAtom(display, "_MOTIF_WM_HINTS", false);
    if (WMHintsAtom) {
      static const unsigned long MWM_HINTS_FUNCTIONS   = 1 << 0;
      static const unsigned long MWM_HINTS_DECORATIONS = 1 << 1;

      //static const unsigned long MWM_DECOR_ALL         = 1 << 0;
      static const unsigned long MWM_DECOR_BORDER      = 1 << 1;
      static const unsigned long MWM_DECOR_RESIZEH     = 1 << 2;
      static const unsigned long MWM_DECOR_TITLE       = 1 << 3;
      //static const unsigned long MWM_DECOR_MENU        = 1 << 4;
      static const unsigned long MWM_DECOR_MINIMIZE    = 1 << 5;
      static const unsigned long MWM_DECOR_MAXIMIZE    = 1 << 6;

      //static const unsigned long MWM_FUNC_ALL          = 1 << 0;
      static const unsigned long MWM_FUNC_RESIZE       = 1 << 1;
      static const unsigned long MWM_FUNC_MOVE         = 1 << 2;
      static const unsigned long MWM_FUNC_MINIMIZE     = 1 << 3;
      static const unsigned long MWM_FUNC_MAXIMIZE     = 1 << 4;
      static const unsigned long MWM_FUNC_CLOSE        = 1 << 5;

      struct WMHints
      {
        unsigned long Flags;
        unsigned long Functions;
        unsigned long Decorations;
        long          InputMode;
        unsigned long State;
      };

      WMHints Hints;
      Hints.Flags       = MWM_HINTS_FUNCTIONS | MWM_HINTS_DECORATIONS;
      Hints.Decorations = 0;
      Hints.Functions   = 0;

      if (true)
      {
        Hints.Decorations |= MWM_DECOR_BORDER | MWM_DECOR_TITLE | MWM_DECOR_MINIMIZE /*| MWM_DECOR_MENU*/;
        Hints.Functions   |= MWM_FUNC_MOVE | MWM_FUNC_MINIMIZE;
      }
      if (false)
      {
        Hints.Decorations |= MWM_DECOR_MAXIMIZE | MWM_DECOR_RESIZEH;
        Hints.Functions   |= MWM_FUNC_MAXIMIZE | MWM_FUNC_RESIZE;
      }
      if (true)
      {
        Hints.Decorations |= 0;
        Hints.Functions   |= MWM_FUNC_CLOSE;
      }

      const unsigned char* HintsPtr = reinterpret_cast<const unsigned char*>(&Hints);
      XChangeProperty(display, nw->window, WMHintsAtom, WMHintsAtom, 32, PropModeReplace, HintsPtr, 5);
    }

    // This is a hack to force some windows managers to disable resizing
    if(true)
    {
      XSizeHints XSizeHints;
      XSizeHints.flags      = PMinSize | PMaxSize;
      XSizeHints.min_width  = XSizeHints.max_width  = width;
      XSizeHints.min_height = XSizeHints.max_height = height;
      XSetWMNormalHints(display, nw->window, &XSizeHints);
    }
  }

  GlopSetTitle(nw, title);
  free((void *)title);

  XSetWMProtocols(display, nw->window, &close_atom, 1);

  nw->inputcontext = XCreateIC(xim, XNInputStyle, XIMPreeditNothing | XIMStatusNothing, XNClientWindow, nw->window, XNFocusWindow, nw->window, NULL);
  if(!nw->inputcontext) {
    LOG_FATAL("couldn't create inputcontext\n");
    abort();
  }

  XMapWindow(display, nw->window);

  glopSetCurrentContext(nw);

  return GlopWindowHandle{nw};
}

GlopWindowHandle DeprecatedGlopCreateWindow(char const* title, int x, int y, int width, int height) {
  OsWindowData *nw = new OsWindowData();

  // this is bad
  if(x == -1) x = 100;
  if(y == -1) y = 100;


  int glxcv_params[] = {
    GLX_RGBA,
    GLX_RED_SIZE, 1,
    GLX_GREEN_SIZE, 1,
    GLX_BLUE_SIZE, 1,
    GLX_DOUBLEBUFFER,
    GLX_DEPTH_SIZE, 1,
    GLX_STENCIL_SIZE, 8,
    None
  };
  nw->vinfo = glXChooseVisual(display, screen, glxcv_params);
  if(!nw->vinfo) {
    LOG_FATAL("couldn't glXChooseVisual\n");
    abort();
  }

  // Define the window attributes
  XSetWindowAttributes attribs;
  attribs.event_mask = KeyPressMask | KeyReleaseMask | ButtonPressMask |
    ButtonReleaseMask | ButtonMotionMask | PointerMotionMask | FocusChangeMask
    | FocusChangeMask | ButtonPressMask | ButtonReleaseMask | ButtonMotionMask
    | PointerMotionMask | KeyPressMask | KeyReleaseMask | StructureNotifyMask |
    EnterWindowMask | LeaveWindowMask;
  attribs.colormap = XCreateColormap( display, RootWindow(display, screen), nw->vinfo->visual, AllocNone);


  nw->window = XCreateWindow(display, RootWindow(display, screen), x, y, width, height, 0, nw->vinfo->depth, InputOutput, nw->vinfo->visual, CWColormap | CWEventMask, &attribs); // I don't know if I need anything further here

  {
    Atom WMHintsAtom = XInternAtom(display, "_MOTIF_WM_HINTS", false);
    if (WMHintsAtom) {
      static const unsigned long MWM_HINTS_FUNCTIONS   = 1 << 0;
      static const unsigned long MWM_HINTS_DECORATIONS = 1 << 1;

      //static const unsigned long MWM_DECOR_ALL         = 1 << 0;
      static const unsigned long MWM_DECOR_BORDER      = 1 << 1;
      static const unsigned long MWM_DECOR_RESIZEH     = 1 << 2;
      static const unsigned long MWM_DECOR_TITLE       = 1 << 3;
      //static const unsigned long MWM_DECOR_MENU        = 1 << 4;
      static const unsigned long MWM_DECOR_MINIMIZE    = 1 << 5;
      static const unsigned long MWM_DECOR_MAXIMIZE    = 1 << 6;

      //static const unsigned long MWM_FUNC_ALL          = 1 << 0;
      static const unsigned long MWM_FUNC_RESIZE       = 1 << 1;
      static const unsigned long MWM_FUNC_MOVE         = 1 << 2;
      static const unsigned long MWM_FUNC_MINIMIZE     = 1 << 3;
      static const unsigned long MWM_FUNC_MAXIMIZE     = 1 << 4;
      static const unsigned long MWM_FUNC_CLOSE        = 1 << 5;

      struct WMHints
      {
        unsigned long Flags;
        unsigned long Functions;
        unsigned long Decorations;
        long          InputMode;
        unsigned long State;
      };

      WMHints Hints;
      Hints.Flags       = MWM_HINTS_FUNCTIONS | MWM_HINTS_DECORATIONS;
      Hints.Decorations = 0;
      Hints.Functions   = 0;

      if (true)
      {
        Hints.Decorations |= MWM_DECOR_BORDER | MWM_DECOR_TITLE | MWM_DECOR_MINIMIZE /*| MWM_DECOR_MENU*/;
        Hints.Functions   |= MWM_FUNC_MOVE | MWM_FUNC_MINIMIZE;
      }
      if (false)
      {
        Hints.Decorations |= MWM_DECOR_MAXIMIZE | MWM_DECOR_RESIZEH;
        Hints.Functions   |= MWM_FUNC_MAXIMIZE | MWM_FUNC_RESIZE;
      }
      if (true)
      {
        Hints.Decorations |= 0;
        Hints.Functions   |= MWM_FUNC_CLOSE;
      }

      const unsigned char* HintsPtr = reinterpret_cast<const unsigned char*>(&Hints);
      XChangeProperty(display, nw->window, WMHintsAtom, WMHintsAtom, 32, PropModeReplace, HintsPtr, 5);
    }

    // This is a hack to force some windows managers to disable resizing
    if(true)
    {
      XSizeHints XSizeHints;
      XSizeHints.flags      = PMinSize | PMaxSize;
      XSizeHints.min_width  = XSizeHints.max_width  = width;
      XSizeHints.min_height = XSizeHints.max_height = height;
      XSetWMNormalHints(display, nw->window, &XSizeHints);
    }
  }

  GlopSetTitle(nw, title);
  free((void *)title);

  XSetWMProtocols(display, nw->window, &close_atom, 1);
  // I think in here is where we're meant to set window styles and stuff

  nw->inputcontext = XCreateIC(xim, XNInputStyle, XIMPreeditNothing | XIMStatusNothing, XNClientWindow, nw->window, XNFocusWindow, nw->window, NULL);
  if(!nw->inputcontext) {
    LOG_FATAL("couldn't create inputcontext\n");
    abort();
  }

  XMapWindow(display, nw->window);

  GLXContext shareList = NULL;
  nw->context = glXCreateContext(display, nw->vinfo, shareList, True);
  if(nw->context == NULL) {
    LOG_FATAL("couldn't create new context\n");
    abort();
  }

  // TODO(tmckee): Use GLFW for window management so that we can do something like
  // glfwSetFramebufferSizeCallback(glfwWindow, setViewportOnResize)
  // Shoutout to @Hermie02 for the suggestion!

  glopSetCurrentContext(nw);

  return GlopWindowHandle{nw};
}


// Destroys a window that is completely or partially created.
void glopDestroyWindow(OsWindowData* data) {
  delete data;
}

void glopGetWindowFocusState(OsWindowData* data, bool* is_in_focus, bool* focus_changed) {
  *is_in_focus = true;
  *focus_changed = false;
}

void glopGetWindowPosition(const OsWindowData* data, int* x, int* y) {
  //XWindowAttributes attrs;
  //XGetWindowAttributes(display, data->window, &attrs);
  // You'd think these functions would do something useful. Problem is, they work relative to the parent window. The parent window is the window that contains the titlebar, nothing more.

  // What we really want to do is to get the absolute offset. The easiest way, as stupid as it is, is to get the cursor position - both relative to window and to world - and subtract.

  // The irony, of course, is that Glop only cares so it can then subtract *again* and get the exact data that we're throwing away right now.

  // mostly ignored
  Window root, child;
  int tx, ty, winx, winy;
  unsigned int mask;
  XQueryPointer(display, data->window, &root, &child, &tx, &ty, &winx, &winy, &mask);

  *x = tx - winx;
  *y = ty - winy;
}

void glopGetWindowSize(const OsWindowData* data, int* width, int* height) {
  XWindowAttributes attrs;
  XGetWindowAttributes(display, data->window, &attrs);
  *width  = attrs.width;
  *height = attrs.height;
}

void GlopGetWindowDims(GlopWindowHandle hdl, int* x, int* y, int* dx, int* dy) {
  glopGetWindowPosition(hdl.data, x, y);
  glopGetWindowSize(hdl.data, dx, dy);
}

void GlopSetWindowSize(GlopWindowHandle hdl, int dx, int dy) {
  // TODO(tmckee): This can generate 'BadValue' or 'BadWindow' errors. We
  // should check for them. See
  // https://tronche.com/gui/x/xlib/event-handling/protocol-errors/XSetErrorHandler.html
  XResizeWindow(display, hdl.data->window, dx, dy);
}

// Input functions
// ===============

void GlopGetInputEvents(GlopWindowHandle hdl, struct GlopKeyEvent** _events_ret, size_t* _num_events, int64_t* _horizon) {
  *_horizon = gt();
  std::vector<struct GlopKeyEvent> ret; // weeeeeeeeeeee
  ret.swap(hdl.data->events);

  *_events_ret = (struct GlopKeyEvent*)malloc(sizeof(struct GlopKeyEvent) * ret.size());
  *_num_events = ret.size();
  for (size_t i = 0; i < ret.size(); i++) {
    // TODO(tmckee): memcpy instead?
    (*_events_ret)[i] = ret[i];
  }
}

// Miscellaneous functions
// =======================

void glopSleep(int t) {
  usleep(t*1000);
}

int glopGetTime() {
  return gt();
}
long long glopGetTimeMicro() {
  return gtm();
}


void GlopSwapBuffers(GlopWindowHandle hdl) {
  glXSwapBuffers(display, hdl.data->window);
}

void GlopEnableVSync(int enable) {
  LOG_WARN("GlopEnableVSync: unimplemented\n");
}

void GlopSetGlContext(GlopWindowHandle hdl) {
  glopSetCurrentContext(hdl.data);
}

} // extern "C"
