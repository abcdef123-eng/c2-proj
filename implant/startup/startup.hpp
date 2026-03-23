#pragma once
#include <iostream>
#include "common.hpp"


BOOL init();
BOOL CollectHostInfo();
BOOL GrabAuthInfo(char* FilePath);
void GenGuid(char* str, size_t len);
const char* is32or64(int num);