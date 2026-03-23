#pragma once
#include "common.hpp"


class bytes {
	

public:
	INT index;
	INT size;
	PBYTE InData;
	PBYTE OutData;

	void Set(PBYTE data, INT DataSize);
	void Write4(BYTE* data, UINT val);
	void WriteBytes(BYTE* data, UINT length, BYTE* string);
	UINT Read4();
	void ReadBytes(BYTE* outData, UINT Lengt);

};

extern bytes* g_Bytes;
