#ifndef AKCHAIN_LIB_H_
#define AKCHAIN_LIB_H_

typedef char* string;
typedef void* JsonRoot;

#define INVOKE_FUNC(func_name, _fn) do{\
    if (strcmp(action, func_name) == 0) {return _fn();}\
}while(0)

#define INVOKE_ACTION(action_name, invoke_) do{\
    if (strcmp(action, action_name) == 0) {invoke_}\
}while(0)


#define EXPORT __attribute__((used))

#define AssertWithReturnBool(var, errStr) \
    if (var == nullptr || var == "") { \
        print_s(errStr); \
        return false; \
    }

#define AssertWithReturnInt(var, errStr) \
    if (var == nullptr || var == "") { \
        print_s(errStr); \
        return -1; \
    }

#define AssertWithReturnString(var, errStr) \
    if (var == nullptr || var == "") { \
        print_s(errStr); \
        return ""; \
    }


#ifdef __cplusplus
extern "C" {
#endif

void print_s(const char *s);
void print_i(int t);
int strlen(const char *s);
int strcmp(const char *s1, const char *s2);
char* strcat(const char *s1, const char *s2);
int Atoi(const char *s);
char* Itoa(int iVal);
char* BigIntSub(const char *bigVal1, const char *bigVal2);
char* BigIntAdd(const char *bigVal1, const char *bigVal2);
int BigIntCmp(const char *bigVal1, const char *bigVal2);
int JsonObjectIndex(const char *s);
int JsonCreateObject(void);
int JsonGetInt(int jsonObjectIndex, const char* argName);
char *JsonGetString(int jsonObjectIndex, const char *argName);
void JsonPutString(int jsonObjectIndex, const char* key, const char* value);
char *JsonToString(int jsonObjectIndex);
int TrigEvent(const char* eventSrc, const char* data);
char *SenderAddr();
char *OwnerAddr();
int ChangeContractOwner(const char* cAddr, const char* ownerAddr);
int SetBalance(const char* addr, const char* symbol, const char* amount);
char* Balance(const char* addr, const char* symbol);
int SetAllowance(const char* addrSender, const char* addrSpender, const char* symbol, const char* amount);
char* Allowance(const char* addrSender, const char* addrSpender, const char* symbol);
char* ContractAddr();
int CreateCurrency(const char* symbol, int decimals, const char* totalSupply);
bool IsContractNormal(const char *cAddr);
int  SuspendContract(const char *cAddr);
int  UnsuspendContract(const char *cAddr);
int BuildCurrencyCAddrMap(const char* symbol, const char* cAddr);
int StoreJsonObject(const char *cAddr, const char *key, const char *jsonObject);
int LoadJsonObject(const char *cAddr, const char *key);
int Height();

#ifdef __cplusplus
}
#endif

#endif/*AKCHAIN_LIB_H_*/