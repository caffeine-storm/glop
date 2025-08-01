#import <Cocoa/Cocoa.h>
#import <OpenGL/gl.h>
#import <glop.h>
#import <mach/mach_time.h>
#import <stdio.h>

#include <ApplicationServices/ApplicationServices.h>
#include <IOKit/hid/IOHIDLib.h>

// TODO: This requires OSX 10.6 or higher, just for getting uptime.
// if we bother to fix linking on osx such that 10.5 is acceptable we
// should change this
#include <Foundation/NSProcessInfo.h>

@interface GlopApplication : NSApplication {
  int should_stop;
  int on_correct_thread;
}
- (int)onCorrectThread;
- (void)sendEvent:(NSEvent*)event;
- (void)stop:(id)id;
- (void)run;
- (void)clear;
@end

struct inputState {
  float mouse_x;
  float mouse_y;

// modifiers
  int num_lock;
  int caps_lock;
  int left_shift;
  int right_shift;
  int left_alt;
  int right_alt;
  int left_ctrl;
  int right_ctrl;
  int left_gui;
  int right_gui;
  int function;
} inputState;

long long NSTimeIntervalToMS(NSTimeInterval t) {
  return (long long)((double)(t) * 1000.0 + 0.5);
}

void ClearEvent(KeyEventOld* event, NSEvent* ns_event) {
  event->timestamp = NSTimeIntervalToMS([ns_event timestamp]);
  event->press_amt = 0;
  event->cursor_x = inputState.mouse_x;
  event->cursor_y = inputState.mouse_y;
  event->num_lock = 0;
  event->caps_lock = 0;
}

NSAutoreleasePool* pool;
NSApplication* glop_app;
NSEvent* terminator;
NSTimeInterval osx_horizon;
CGPoint lock_mouse;

// These structures provide a way to allow threads to write events to a buffer
// and then grab the events as a batch in a synchronously.
// TODO: Would be nice to have an auto-expanding array here
typedef struct {
  KeyEventOld events[1000];
  int length;
} EventGroup;

EventGroup event_buffer_1;
EventGroup event_buffer_2;
EventGroup *current_event_buffer;
pthread_mutex_t event_group_mutex;

// Safely adds the event to current event buffer and increments its length
void AddEvent(KeyEventOld* event) {
  pthread_mutex_lock(&event_group_mutex);

  current_event_buffer->events[current_event_buffer->length] = *event;
  current_event_buffer->length++;

  pthread_mutex_unlock(&event_group_mutex);
}

// Returns a pointer to all of the current events as well as the number of events in the buffer
// Swaps the current buffer so that new events go into the other buffer, the events returned
// by this function should be used before the next time this function is called.
void GetEventsOld(KeyEventOld** events, int* length, long long* horizon) {
  pthread_mutex_lock(&event_group_mutex);

  *events = current_event_buffer->events;
  *length = current_event_buffer->length;

  if (current_event_buffer == &event_buffer_1) {
    current_event_buffer = &event_buffer_2;
  } else {
    current_event_buffer = &event_buffer_1;
  }
  current_event_buffer->length = 0;

  *horizon = NSTimeIntervalToMS(osx_horizon);
  pthread_mutex_unlock(&event_group_mutex);
}

typedef struct {
  IOHIDDeviceRef* devices;
  int num_devices;
  int max_devices;
  IOHIDManagerRef manager;
  pthread_mutex_t mutex;
} glopHidManagerStruct;
glopHidManagerStruct glop_hid_manager;
void hidCallbackInsert(
    void* context,
    IOReturn result,
    void* sender,
    IOHIDDeviceRef device) {
  pthread_mutex_lock(&glop_hid_manager.mutex);
  if (glop_hid_manager.num_devices == glop_hid_manager.max_devices) {
    glop_hid_manager.max_devices *= 2;
    glop_hid_manager.devices =
        (IOHIDDeviceRef*)realloc(
            glop_hid_manager.devices,
            sizeof(IOHIDDeviceRef*) * glop_hid_manager.max_devices);
    }
  glop_hid_manager.devices[glop_hid_manager.num_devices] = device;
  glop_hid_manager.num_devices++;
  pthread_mutex_unlock(&glop_hid_manager.mutex);
}
void hidCallbackRemove(
    void* context,
    IOReturn result,
    void* sender,
    IOHIDDeviceRef device) {
  pthread_mutex_lock(&glop_hid_manager.mutex);
  int i;
  for (i = 0; i < glop_hid_manager.num_devices; i++) {
    if (glop_hid_manager.devices[i] == device) {
      glop_hid_manager.devices[i] = glop_hid_manager.devices[glop_hid_manager.num_devices - 1];
      glop_hid_manager.num_devices--;
      break;
    }
  }
  pthread_mutex_unlock(&glop_hid_manager.mutex);
}
void initGlopHidManager() {
  // Make an IOHID manager and get it ready to respond to HID plugins and
  // removals.
  glop_hid_manager.manager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
  if (glop_hid_manager.manager == NULL) {
    printf("Failed to init IOHID manager.\n");
    return;
  }
  pthread_mutex_init(&glop_hid_manager.mutex, NULL);
  glop_hid_manager.devices = (IOHIDDeviceRef*)malloc(sizeof(IOHIDDeviceRef*) * 10);
  glop_hid_manager.num_devices = 0;
  glop_hid_manager.max_devices = 10;

  // Match everything, because why not?
  IOHIDManagerSetDeviceMatching(glop_hid_manager.manager, NULL);
  IOHIDManagerRegisterDeviceMatchingCallback(glop_hid_manager.manager, hidCallbackInsert, NULL);
  IOHIDManagerRegisterDeviceRemovalCallback(glop_hid_manager.manager, hidCallbackRemove, NULL);

  // Now open the IO HID Manager reference
  IOReturn io_return = IOHIDManagerOpen(glop_hid_manager.manager, kIOHIDOptionsTypeNone);
  if (io_return != 0) {
    printf("Failed to open IOHID manager.\n");
  }
  IOHIDManagerScheduleWithRunLoop(glop_hid_manager.manager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
}

void Init() {
  glop_app = [GlopApplication sharedApplication];
  [(GlopApplication*)glop_app clear];
  pool = [[NSAutoreleasePool alloc] init];

  terminator = [NSEvent
      otherEventWithType:NSApplicationDefined
      location:NSZeroPoint
      modifierFlags:0
      timestamp:(NSTimeInterval)0
      windowNumber:0
      context:0
      subtype:0
      data1:0
      data2:0];

  // EventGroup init stuff
  current_event_buffer = &event_buffer_1;
  event_buffer_1.length = 0;
  event_buffer_2.length = 0;
  osx_horizon = [[NSProcessInfo processInfo] systemUptime];
  pthread_mutex_init(&event_group_mutex, NULL);
  lock_mouse.x = -1;

  initGlopHidManager();

  [glop_app finishLaunching];
}


// This is a map of keycodes generated by OSX to Glop keycodes
// This does not contain any modifier keys, those are handled separately
const int key_map[] = {
  'a', 's', 'd', 'f', 'h',
  'g', 'z', 'x', 'c', 'v',
    0, 'b', 'q', 'w', 'e',
  'r', 'y', 't', '1', '2',
  '3', '4', '6', '5', '=',
  '9', '7', '-', '8', '0',
  ']', 'o', 'u', '[', 'i',
  'p', kKeyEnter, 'l', 'j', '\'',
  'k', ';', '\\', ',', '/',
  'n', 'm', '.', kKeyTab, ' ',
  '`', kKeyBackspace, 0, kKeyEscape, 0,             // 50
  0, 0, 0, 0, 0,
//  kKeyGUI, 0, 0, 0, 0,   <- TODO: Where is this supposed to be defined?
  0, 0, 0, 0, 0,
  kKeyPadDecimal, 0, kKeyPadMultiply, 0, kKeyPadAdd,
  0, 0, 0, 0, 0,
//  0, kKeyPadClear, 0, 0, 0,  <- TODO: Where is this supposed to be defined?
  kKeyPadDivide, 0, kKeyPadEnter, kKeyPadSubtract, 0,
  0, 0, kKeyPad0, kKeyPad1, kKeyPad2,
  kKeyPad3, kKeyPad4, kKeyPad5, kKeyPad6, kKeyPad7,
  kKeyPad8, kKeyPad9, 0, 0, 0,
  0, kKeyF5, kKeyF6, kKeyF7, kKeyF3,
  kKeyF8, kKeyF9, 0, kKeyF11, 0,                    // 100
  kKeyPrintScreen, 0, 0, 0, kKeyF10,
  0, kKeyF12, 0, 0, 0,
//  0, kKeyF12, 0, 0, kKeyHelp,  <- TODO: Where is this supposed to be defined?
  kKeyHome, kKeyPageUp, kKeyDelete, kKeyF4, kKeyEnd,
  kKeyF2, kKeyPageDown, kKeyF1, kKeyLeft, kKeyRight,
  kKeyDown, kKeyUp,
  -1, -1, -1, -1, -1
};

// Modifier flags
const int kOsxCapsLock =      0x10000;
const int kOsxFunction =     0x800000;
const int kOsxLeftControl =   0x40001;
// TODO: Use a keyboard with a Right Control button so we can make sure we get this value right!!
//       Num lock too!!!
const int kOsxRightControl =  0x40000;
const int kOsxLeftShift =     0x20002;
const int kOsxRightShift =    0x20004;
const int kOsxLeftAlt =       0x80020;
const int kOsxRightAlt =      0x80040;
const int kOsxLeftGui =      0x100008;
const int kOsxRightGui =     0x100010;

int* getInputStateVal(int flag) {
  if (flag == kOsxCapsLock)     return &inputState.caps_lock;
  if (flag == kOsxFunction)     return &inputState.function;
  if (flag == kOsxLeftShift)    return &inputState.left_shift;
  if (flag == kOsxRightShift)   return &inputState.right_shift;
  if (flag == kOsxLeftAlt)      return &inputState.left_alt;
  if (flag == kOsxRightAlt)     return &inputState.right_alt;
  if (flag == kOsxLeftControl)  return &inputState.left_ctrl;
  if (flag == kOsxRightControl) return &inputState.right_ctrl;
  if (flag == kOsxLeftGui)      return &inputState.left_gui;
  if (flag == kOsxRightGui)     return &inputState.right_gui;
  return 0;
}

@implementation GlopApplication
- (void)clear {
  should_stop = 0;
  on_correct_thread = 0;
}

- (void)sendEvent:(NSEvent*)event {
/*
   NSLeftMouseDown      = 1,
   NSLeftMouseUp        = 2,
   NSRightMouseDown     = 3,
   NSRightMouseUp       = 4,
   NSMouseMoved         = 5,
   NSKeyDown            = 10,
   NSKeyUp              = 11,
   NSFlagsChanged       = 12,
   NSApplicationDefined = 15,
   NSPeriodic           = 16,
   NSCursorUpdate       = 17,
   NSScrollWheel        = 22,
   NSOtherMouseDown     = 25,
   NSOtherMouseUp       = 26,
*/
// TODO: We need this here if we want to be able to break out of the Run loop
  if ([event type] == NSApplicationDefined) {
    [glop_app stop:self];
    return;
  }
  if ([event type] == NSFlagsChanged) {
    NSUInteger flags =  [event modifierFlags];
    int flag[8];
    flag[0] = kOsxCapsLock;
    flag[1] = kOsxLeftControl;
    flag[2] = kOsxLeftShift;
    flag[3] = kOsxRightShift;
    flag[4] = kOsxLeftAlt;
    flag[5] = kOsxRightAlt;
    flag[6] = kOsxLeftGui;
    flag[7] = kOsxRightGui;
    int glopKeyCode[8];
    glopKeyCode[0] = kKeyCapsLock;
    glopKeyCode[1] = kKeyLeftControl;
    glopKeyCode[2] = kKeyLeftShift;
    glopKeyCode[3] = kKeyRightShift;
    glopKeyCode[4] = kKeyLeftAlt;
    glopKeyCode[5] = kKeyRightAlt;
    glopKeyCode[6] = kKeyLeftGui;
    glopKeyCode[7] = kKeyRightGui;
    int i;
    for (i = 0; i < 8; i++) {
      int* val = getInputStateVal(flag[i]);
      if ((*val != 0) != ((flag[i] & flags) == flag[i])) {
        KeyEventOld key_event;
        ClearEvent(&key_event, event);
        key_event.index = glopKeyCode[i];
        if (*val == 0) {
          *val = 1;
        } else {
          *val = 0;
        }
        key_event.press_amt = *val;
        AddEvent(&key_event);
      }
    }
  } else if ([event type] == NSScrollWheel) {
    NSPoint cursor_pos = [event locationInWindow];
    NSWindow* window = [event window];
    if (window != nil) {
      NSRect rect;
      cursor_pos = [window convertBaseToScreen:cursor_pos];
    }
    inputState.mouse_x = cursor_pos.x;
    inputState.mouse_y = cursor_pos.y;
    KeyEventOld scroll_event;
    ClearEvent(&scroll_event, event);
    scroll_event.press_amt = [event deltaY];
    scroll_event.index = kMouseWheelVertical;
    if (scroll_event.press_amt != 0) {
      AddEvent(&scroll_event);
    }
  } else if ([event type] == NSMouseMoved ||
             [event type] == NSLeftMouseDragged ||
             [event type] == NSRightMouseDragged ||
             [event type] == NSOtherMouseDragged) {
    // TODO: It looks like OSX will only give us one MouseMoved event per Think, it
    // must be modifyin whatever MouseMoved event is in the queue as new MouseMoved
    // events come in.  To get better resolution we need to find the cursor position
    // when other events happen and generate mouse moved events for each one.
    NSPoint cursor_pos = [event locationInWindow];
    NSWindow* window = [event window];
    if (window != nil) {
      NSRect rect;
      cursor_pos = [window convertBaseToScreen:cursor_pos];
    }
    inputState.mouse_x = cursor_pos.x;
    inputState.mouse_y = cursor_pos.y;
    KeyEventOld key_x_event;
    ClearEvent(&key_x_event, event);
    key_x_event.index = kMouseXAxis;
    key_x_event.press_amt = [event deltaX];
    key_x_event.cursor_x = inputState.mouse_x;
    if (key_x_event.press_amt != 0) {
      AddEvent(&key_x_event);
    }

    KeyEventOld key_y_event;
    ClearEvent(&key_y_event, event);
    key_y_event.index = kMouseYAxis;
    key_y_event.press_amt = [event deltaY];
    key_y_event.cursor_y = inputState.mouse_y;
    if (key_y_event.press_amt != 0) {
      AddEvent(&key_y_event);
    }
  } else if ([event type] == NSKeyDown ||
             [event type] == NSKeyUp) {
    KeyEventOld key_event;
    ClearEvent(&key_event, event);
    key_event.index = key_map[[event keyCode]];
    key_event.press_amt = 0;
    if ([event type] == NSKeyDown) {
      key_event.press_amt = 1;
    }
    AddEvent(&key_event);
  } else if ([event type] == NSLeftMouseDown  ||
             [event type] == NSLeftMouseUp    ||
             [event type] == NSRightMouseDown ||
             [event type] == NSRightMouseUp) {
    KeyEventOld key_event;
    ClearEvent(&key_event, event);
    key_event.index = -1;
    if ([event type] == NSLeftMouseDown || [event type] == NSLeftMouseUp) {
      key_event.index = kMouseLButton;
    }
    if ([event type] == NSRightMouseDown || [event type] == NSRightMouseUp) {
      key_event.index = kMouseRButton;
    }
    key_event.press_amt = 0;
    if ([event type] == NSLeftMouseDown || [event type] == NSRightMouseDown) {
      key_event.press_amt = 1;
    }
    AddEvent(&key_event);
  } else {
    [super sendEvent: event];
  }
}

- (void)stop:(id)id {
  should_stop = 1;
}

- (int)onCorrectThread {
  return on_correct_thread;
}

- (void)run {
//  NSAutoreleasePool *pool = [[NSAutoreleasePool alloc] init];

  do {
//    [pool release];
//    pool = [[NSAutoreleasePool alloc] init];
    NSEvent *event =
      [self
        nextEventMatchingMask:NSAnyEventMask
        untilDate:[NSDate distantFuture]
        inMode:NSDefaultRunLoopMode
        dequeue:YES];

    on_correct_thread = ([event type] != 0);
    if (!on_correct_thread) {
      return;
    }
    [self sendEvent:event];
    [self updateWindows];
  } while (!should_stop);
  should_stop = 0;
//  [pool release];
}
@end

void Quit() {
  [glop_app postEvent:terminator atStart:FALSE];
}

int Think() {
  // TODO: This is retarded, but it does seem to get all of the evnts out of the queue
  // rather than only most of them
  [glop_app postEvent:terminator atStart:FALSE];
  [glop_app run];
  if (![(GlopApplication*)glop_app onCorrectThread]) {
    return 0;
  }
  [glop_app postEvent:terminator atStart:FALSE];
  [glop_app run];
  if (![(GlopApplication*)glop_app onCorrectThread]) {
    return 0;
  }
  osx_horizon = [[NSProcessInfo processInfo] systemUptime];
  if (lock_mouse.x >= 0) {
    CGWarpMouseCursorPosition(lock_mouse);
  }
  return 1;
}

void GetInputEvents(void** _key_events, int* length, long long* horizon) {
  // Loop over all IOHID queues and turn each event into a keyevent
  // GetEvents((KeyEventOld**)_key_events, length, horizon);
  *length = 0;
  *horizon = 0;
}

void GetInputEventsOld(void** _key_events, int* length, long long* horizon) {
  GetEventsOld((KeyEventOld**)_key_events, length, horizon);
}

void CreateWindow(void** _window, void** _context, int x, int y, int width, int height) {
  NSRect windowRect = NSMakeRect(x, y, width, height);
  NSWindow* window = [NSWindow alloc];
  *((NSWindow**)(_window)) = window;
  [window initWithContentRect:windowRect
  styleMask:( NSResizableWindowMask | NSClosableWindowMask | NSTitledWindowMask)
  backing:NSBackingStoreBuffered defer:NO];
  [window makeKeyAndOrderFront:nil];
  [window setAcceptsMouseMovedEvents:YES];
  NSPoint window_cursor = [window mouseLocationOutsideOfEventStream];
  NSPoint cursor = [window convertBaseToScreen:window_cursor];
  inputState.mouse_x = cursor.x;
  inputState.mouse_y = cursor.y;

  // Create and bind an OpenGL context
  NSOpenGLPixelFormatAttribute attributes[] = {
    NSOpenGLPFADoubleBuffer,
    NSOpenGLPFAAccelerated,
    NSOpenGLPFAColorSize, 32,
    NSOpenGLPFADepthSize, 32,
    NSOpenGLPFAStencilSize, 8,
    //    NSOpenGLPFAFullScreen,
    0,
  };
  NSOpenGLPixelFormat* pixel_format = [[NSOpenGLPixelFormat alloc] initWithAttributes:attributes];
  if (pixel_format == nil) {
    // TODO: How do we signal this properly?
    exit(0);
    return;
  }
  NSOpenGLContext* context = [NSOpenGLContext alloc];
  *((NSOpenGLContext**)(_context)) = context;
  [context initWithFormat:pixel_format shareContext:NO];
  [context setView:[window contentView]];
  [context makeCurrentContext];
  glClear(GL_COLOR_BUFFER_BIT);
  [context flushBuffer];
}

void SwapBuffers(void* _context) {
  NSOpenGLContext* context = (NSOpenGLContext*)(_context);
  [context flushBuffer];
}

void ShutDown() {
  [pool drain];
}

void Run() {
  [NSApp run];
}

void GetMousePos(int* x, int* y) {
  NSPoint point = [NSEvent mouseLocation];
  *x = (int)point.x;
  *y = (int)point.y;
}

void LockCursor(int lock) {
  if (lock) {
    CGEventRef dummy = CGEventCreate(NULL);
    lock_mouse = CGEventGetLocation(dummy);
  } else {
    lock_mouse.x = -1;
  }
}

void HideCursor(int hide) {
  if (hide) {
    CGDisplayHideCursor(kCGDirectMainDisplay);
  } else {
    CGDisplayShowCursor(kCGDirectMainDisplay);
  }
}

void GetWindowDims(void* _window, int* x, int* y, int* dx, int* dy) {
  NSWindow* window = (NSWindow*)_window;
  NSRect view = [[window contentView] frame];
  NSRect rect = [window frame];
  *x = rect.origin.x;
  *y = rect.origin.y;
  *dx = view.size.width;
  *dy = view.size.height;
}

void EnableVSync(void* _context, int set_vsync) {
  NSOpenGLContext* context = (NSOpenGLContext*)(_context);
  GLint swapInt = set_vsync;
  [context setValues:&swapInt forParameter:NSOpenGLCPSwapInterval];
}

