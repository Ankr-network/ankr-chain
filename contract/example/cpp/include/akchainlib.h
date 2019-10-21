#ifndef ANCHAIN_LIB_H_
#define ANCHAIN_LIB_H_

typedef char* string;
typedef void* JsonRoot;

#define INVOKE_FUNC(func_name, _fn) do{\
    if (strcmp(action, func_name) == 0) {return _fn();}\
}while(0)

#define INVOKE_ACTION(action_name, invoke_) do{\
    if (strcmp(action, action_name) == 0) {invoke_}\
}while(0)


#define EXPORT __attribute__((used))

#ifdef __cplusplus
extern "C" {
#endif

void print_s(const char *s);
void print_i(int t);
int strlen(const char *s);
int strcmp(const char *s1, const char *s2);
int JsonObjectIndex(const char *s);
int JsonCreateObject(void);
int JsonGetInt(int jsonObjectIndex, const char* argName);
char *JsonGetString(int jsonObjectIndex, const char *argName);
void JsonPutString(int jsonObjectIndex, const char* key, const char* value);
char *JsonToString(int jsonObjectIndex);
int TrigEvent(const char* eventSrc, const char* data);

#ifdef __cplusplus
}
#endif

#endif/*ANCHAIN_LIB_H_*/