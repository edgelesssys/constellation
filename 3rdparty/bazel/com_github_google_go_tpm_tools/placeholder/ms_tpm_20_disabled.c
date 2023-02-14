#include "Platform.h"
#include "Tpm.h"

int g_inFailureMode = 0;

void _plat__Reset(bool forceManufacture) {}

void _plat__RunCommand(uint32_t requestSize, unsigned char *request,
                       uint32_t *responseSize, unsigned char **response) {}
