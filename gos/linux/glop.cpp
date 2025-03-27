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

using namespace std;

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
  vector<struct GlopKeyEvent> events;
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
      fprintf(stderr, "fatal: couldn't open X display\n");
      abort();
    }

    screen = DefaultScreen(display);

    xim = XOpenIM(display, NULL, NULL, NULL);
    if(xim == NULL) {
      fprintf(stderr, "fatal: couldn't open X input method\n");
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

static bool SynthKey(KeySym const &sym, bool pushed, XKeyEvent const &event, Window window, struct GlopKeyEvent *ev) {
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
    case XK_space: ki = '/'; break;
  }

  if(ki == 0)
    return false;

  ev->index = ki;
  ev->device_type = glopDeviceKeyboard;
  ev->press_amt = pushed ? 1.0 : 0.0;
  ev->timestamp = gt();
  ev->cursor_x = event.x;
  ev->cursor_y = event.y;
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

static bool SynthButton(bool pushed, XButtonEvent const &event, Window window, struct GlopKeyEvent *ev) {

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
    case Button4:
      ki = kMouseWheelUp;
      break;
    case Button5:
      ki = kMouseWheelDown;
      break;
    default:
      fprintf(stderr, "SynthButton: unknown button: %d\n", event.button);
      return false;
  }

  ev->index = ki;
  ev->device_type = glopDeviceMouse;
  ev->press_amt = pushed ? 1.0 : 0.0;
  ev->timestamp = gt();
  ev->cursor_x = event.x;
  ev->cursor_y = event.y;
  ev->num_lock = event.state & (1 << 4);
  ev->caps_lock = event.state & LockMask;
  return true;
}

static bool SynthMotion(int dx, int dy, const XMotionEvent &event, Window window, struct GlopKeyEvent *ev, struct GlopKeyEvent *ev2) {
  ev->index = kMouseXAxis;
  ev->device_type = glopDeviceMouse;
  ev->press_amt = dx;
  ev->timestamp = gt();
  ev->cursor_x = event.x;
  ev->cursor_y = event.y;
  ev->num_lock = event.state & (1 << 4);
  ev->caps_lock = event.state & LockMask;

  ev2->index = kMouseYAxis;
  ev2->device_type = glopDeviceMouse;
  ev2->press_amt = dy;
  ev2->timestamp = ev->timestamp;
  ev2->cursor_x = event.x;
  ev2->cursor_y = event.y;
  ev2->num_lock = event.state & (1 << 4);
  ev2->caps_lock = event.state & LockMask;

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

        if(SynthKey(sym, event.type == KeyPress, event.xkey, data->window, &ev))
          data->events.push_back(ev);
        break;
      }

      case ButtonPress:
      case ButtonRelease:
        if(SynthButton(event.type == ButtonPress, event.xbutton, data->window, &ev))
          data->events.push_back(ev);
        break;

      case MotionNotify: {
        struct GlopKeyEvent ev2;
        GlopClearKeyEvent(&ev2);
        if(SynthMotion(event.xmotion.x, event.xmotion.y, event.xmotion, data->window, &ev, &ev2)) {
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
          // TODO: probably want to do something here
//        WindowDashDestroy(); // ffffff
//        LOGF("destroed\n");
        return gt();

      case ClientMessage :
        // TODO(tmckee#22): ummm... shouldn't we close the window? we should
        // verify that this case runs if someone clicks the 'x' through
        // window-decoration or w/e
        if(event.xclient.format == 32 && event.xclient.data.l[0] == static_cast<long>(close_atom)) {
//            WindowDashDestroy();
//            LOGF("destroj\n");
            return gt();
        }
    }
  }

  return gt();
}

void GlopSetTitle(OsWindowData* data, const string& title) {
  XStoreName(display, data->window, title.c_str());
}

void glopSetCurrentContext(OsWindowData* data) {
  if(!glXMakeCurrent(display, data->window, data->context)) {
    fprintf(stderr, "glop.cpp: glxMakeCurrent failed\n");
    exit(1);
  }
}

GlopWindowHandle GlopCreateWindow(char const* title, int x, int y, int width, int height) {
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
//  ASSERT(nw->vinfo);

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
//  ASSERT(nw->inputcontext);

  XMapWindow(display, nw->window);

  nw->context = glXCreateContext(display, nw->vinfo, NULL, True);
//  ASSERT(nw->context);

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
  vector<struct GlopKeyEvent> ret; // weeeeeeeeeeee
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
  fprintf(stderr, "WARN: glop.cpp: GlopEnableVSync: unimplemented\n");
}

void GlopSetGlContext(GlopWindowHandle hdl) {
  glopSetCurrentContext(hdl.data);
}

} // extern "C"
