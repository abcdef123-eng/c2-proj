#pragma once
#include <Windows.h>





//const char* Domain              = "https://192.168.1.24";
//const int port                  = 443;
//const char* GetEndpoint         = "/get";
//const char* registerEndpoint    = "/register";
//const char* postEndpoint        = "/post";


typedef struct _AuthConfig{
	PCHAR guid;
	PCHAR authToken;
	PCHAR RefreshToken;
} AuthConfig, *PAuthConfig;


typedef struct _CommandInfo {
	UINT commandCode;
	PCHAR param;
	PCHAR param2;
	UINT commandID;
} COMMANDINFO, *PCOMMANDINFO;

typedef struct _ServerConfig {
	char BaseURL[128];
	char postEndpoint[32];
	char getEndpoint[32];
	char registerEndpoint[32];
	INT Port;
	INT   UseSSL;
	INT SleepTime;
	INT Jitter;
} ServerConfig, * PServerConfig;


extern PAuthConfig AuthConf;
extern PServerConfig ServerConf;
extern PCOMMANDINFO CommandInfo;



void XOR(PCHAR src, SIZE_T size, PCHAR out, UCHAR key);
BOOL ParseCommand(BYTE* data, INT SizeOfData);

