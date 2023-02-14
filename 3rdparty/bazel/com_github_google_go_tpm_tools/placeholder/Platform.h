#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

extern int g_inFailureMode;

typedef union {
	uint16_t      size;
	uint8_t       *buffer;
} TPM2B, TPM2B_SEED;
typedef struct
{
    TPM2B_SEED          EPSeed;
    TPM2B_SEED          SPSeed;
    TPM2B_SEED          PPSeed;
} PERSISTENT_DATA;

extern PERSISTENT_DATA  gp;

void _plat__Reset(bool forceManufacture);
void _plat__RunCommand(uint32_t requestSize, unsigned char *request,
                       uint32_t *responseSize, unsigned char **response);
