#ifndef GLOP_GOS_LINUX_GLOP_H
#define GLOP_GOS_LINUX_GLOP_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// GlopKey devices
#define glopDeviceKeyboard -1
#define glopDeviceMouse -2
#define glopDeviceDerived -3
#define glopMinDevice -3

#define kAnyKey -1
#define kNoKey -2

#define kKeyBackspace 8
#define kKeyTab 9
#define kKeyEnter 13
#define kKeyReturn 13
#define kKeyEscape 27

#define kKeyF1 129
#define kKeyF2 130
#define kKeyF3 131
#define kKeyF4 132
#define kKeyF5 133
#define kKeyF6 134
#define kKeyF7 135
#define kKeyF8 136
#define kKeyF9 137
#define kKeyF10 138
#define kKeyF11 139
#define kKeyF12 140

#define kKeyCapsLock 150
#define kKeyNumLock 151
#define kKeyScrollLock 152
#define kKeyPrintScreen 153
#define kKeyPause 154
#define kKeyLeftShift 155
#define kKeyRightShift 156
#define kKeyLeftControl 157
#define kKeyRightControl 158
#define kKeyLeftAlt 159
#define kKeyRightAlt 160
#define kKeyLeftGui 161
#define kKeyRightGui 162

#define kKeyRight 166
#define kKeyLeft 167
#define kKeyUp 168
#define kKeyDown 169

#define kKeyPadDivide 170
#define kKeyPadMultiply 171
#define kKeyPadSubtract 172
#define kKeyPadAdd 173
#define kKeyPadEnter 174
#define kKeyPadDecimal 175
#define kKeyPadEquals 176
#define kKeyPad0 177
#define kKeyPad1 178
#define kKeyPad2 179
#define kKeyPad3 180
#define kKeyPad4 181
#define kKeyPad5 182
#define kKeyPad6 183
#define kKeyPad7 184
#define kKeyPad8 185
#define kKeyPad9 186

#define kKeyDelete 190
#define kKeyHome 191
#define kKeyInsert 192
#define kKeyEnd 193
#define kKeyPageUp 194
#define kKeyPageDown 195

#define kMouseXAxis 300
#define kMouseYAxis 301
#define kMouseWheelVertical 302
#define kMouseWheelHorizontal 303
#define kMouseLButton 304
#define kMouseRButton 305
#define kMouseMButton 306

struct GlopKeyEvent {
  short index;
  short device_type;
  float press_amt;
  long long timestamp;

  // X and Y co-ordinates of the mouse at the time the event happened.  In
  // units of pixels with the bottom-most, left-most pixel as the origin.
  int cursor_x;
  int cursor_y;

  int num_lock;
  int caps_lock;
};

void GlopClearKeyEvent(struct GlopKeyEvent* event);

struct OsWindowData;
typedef struct {
  struct OsWindowData* data;
} GlopWindowHandle;
unsigned long GetNativeHandle(GlopWindowHandle);

int64_t GlopInit();
// Returns an opaque handle for further window operations.
GlopWindowHandle GlopCreateWindowHandle(char const* title, int x, int y,
                                        int width, int height);
// Returns an opaque handle for further window operations.
GlopWindowHandle DeprecatedGlopCreateWindow(char const* title, int x, int y,
                                            int width, int height);

// Returns the current time like GetInputEvents' |_horizon|.
int64_t GlopThink(GlopWindowHandle);
void GlopSwapBuffers(GlopWindowHandle);

void GlopGetWindowDims(GlopWindowHandle, int* x, int* y, int* dx, int* dy);
void GlopSetWindowSize(GlopWindowHandle, int dx, int dy);
// The caller is responsible for calling free(*_events_ret)
void GlopGetInputEvents(GlopWindowHandle, struct GlopKeyEvent** _events_ret,
                        size_t* _num_events, int64_t* _horizon);
void GlopEnableVSync(int enable);

// Takes a handle returned from GlopCreateWindow.
void GlopSetGlContext(GlopWindowHandle hdl);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // GLOP_GOS_LINUX_GLOP_H
