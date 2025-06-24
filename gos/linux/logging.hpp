#ifndef GLOP_GOS_LOGGING_HPP
#define GLOP_GOS_LOGGING_HPP

#include <iostream>

#define LOGGING_LEVEL_FATAL 4
#define LOGGING_LEVEL_ERROR 3
#define LOGGING_LEVEL_WARN 2
#define LOGGING_LEVEL_DEBUG 1

// By default, only DEBUG messages are suppressed
#define LOGGING_LEVEL LOGGING_LEVEL_WARN

#define DO_LOG(lvl, expr) \
  std::cerr << __FILE__ ":" << __LINE__ << " " lvl ": " << expr << std::endl

#if LOGGING_LEVEL <= LOGGING_LEVEL_FATAL
#define LOG_FATAL(expr) DO_LOG("FATAL", expr)
#else
#define LOG_FATAL(expr) \
  do {                  \
  } while (false)
#endif  // LOGGING_LEVEL <= LOGGING_LEVEL_FATAL

#if LOGGING_LEVEL <= LOGGING_LEVEL_ERROR
#define LOG_ERROR(expr) DO_LOG("ERROR", expr)
#else
#define LOG_ERROR(expr) \
  do {                  \
  } while (false)
#endif  // LOGGING_LEVEL <= LOGGING_LEVEL_ERROR

#if LOGGING_LEVEL <= LOGGING_LEVEL_WARN
#define LOG_WARN(expr) DO_LOG("WARN", expr)
#else
#define LOG_WARN(expr) \
  do {                 \
  } while (false)
#endif  // LOGGING_LEVEL <= LOGGING_LEVEL_WARN

#if LOGGING_LEVEL <= LOGGING_LEVEL_DEBUG
#define LOG_DEBUG(expr) DO_LOG("DEBUG", expr)
#else
#define LOG_DEBUG(expr) \
  do {                  \
  } while (false)
#endif  // LOGGING_LEVEL <= LOGGING_LEVEL_DEBUG

#endif  // GLOP_GOS_LOGGING_HPP
