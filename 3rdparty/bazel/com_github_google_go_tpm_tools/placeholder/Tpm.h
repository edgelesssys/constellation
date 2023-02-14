#undef TRUE
#define TRUE                1
#undef FALSE
#define FALSE               0
#undef YES
#define YES                 1
#undef NO
#define NO                  0
#undef SET
#define SET                 1
#undef CLEAR
#define CLEAR               0
#ifndef MAX_RESPONSE_SIZE
#define MAX_RESPONSE_SIZE               4096
#endif

#ifndef EPSeed
#define EPSeed 1
#endif
#ifndef SPSeed
#define SPSeed 1
#endif
#ifndef PPSeed
#define PPSeed 1
#endif

#define NV_SYNC_PERSISTENT(x)
