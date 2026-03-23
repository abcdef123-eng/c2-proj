#include "bytes.hpp"
#include <iostream>


/*
	TODO
		handle bounds checks 
		
*/


void bytes::Set(PBYTE data, INT DataSize) {
	this->index = 0;
	this->size = DataSize;
	this->InData = data;
}




void bytes::Write4(BYTE* data, UINT val) {
	memcpy(data + index, &val, sizeof(val));
	this->index += 4;
}


void bytes::WriteBytes(BYTE* data, UINT length, BYTE* string) {
	memcpy(data + index, string, length);
	this->index += length;
}

UINT bytes::Read4() {
	UINT val;
	memcpy(&val, this->InData + this->index, sizeof(val));
	this->index += 4;

	return val;
}
void bytes::ReadBytes(BYTE* outData, UINT Length) {
	memcpy(outData, this->InData + this->index, Length);

	this->index += Length;

}


bytes* g_Bytes = NULL;
