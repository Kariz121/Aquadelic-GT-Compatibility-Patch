/*
    Aquadelic GT Compatibility Patch
    --------------------------------
    Unofficial community compatibility patcher.

    This program does NOT contain the complete original or modified Run.exe.
    It identifies one supported original executable by SHA-256 and applies
    only the required replacement bytes.

    Supported original SHA-256:
    36fcf10bd35f0ea3844b16d6ecc8ac67752ce878f7a40ace5192d91d2017d89a

    Patched SHA-256:
    949768e738e43aa2240176b752ac2aad9d0b66dc442aa87fd21634b58d738aa5

    Target: Win32 / Windows XP (5.1) and newer.
*/

#define WIN32_LEAN_AND_MEAN
#define WINVER 0x0501
#define _WIN32_WINNT 0x0501

#include <windows.h>
#include <commdlg.h>
#include <stdio.h>
#include <stdint.h>
#include <string.h>

#define APP_TITLE L"Aquadelic GT Patcher"
#define BACKUP_DIR L"Original files before patch"
#define MAX_GAME_PATH 32768

static const char *ORIGINAL_SHA256 =
    "36fcf10bd35f0ea3844b16d6ecc8ac67752ce878f7a40ace5192d91d2017d89a";
static const char *PATCHED_SHA256 =
    "949768e738e43aa2240176b752ac2aad9d0b66dc442aa87fd21634b58d738aa5";

typedef struct {
    DWORD offset;
    DWORD size;
    const BYTE *data;
} PatchRegion;

static const BYTE patch_0[] = { 0x00, 0x10 };
static const BYTE patch_1[] = { 0xB0, 0x2E, 0x90, 0x90, 0x90 };
static const BYTE patch_2[] = { 0xE8, 0x0C, 0xA8, 0x17, 0x00, 0xE9, 0xBB, 0xFF, 0xFF, 0xFF };
static const BYTE patch_3[] = { 0xE9, 0xB3, 0x91, 0x14, 0x00, 0x90, 0x90, 0x90 };
static const BYTE patch_4[] = { 0xCF, 0x44, 0x0C, 0x00 };
static const BYTE patch_5[] = { 0x67, 0x43, 0x0C, 0x00 };
static const BYTE patch_6[] = { 0xBB, 0x3D, 0x0C, 0x00 };
static const BYTE patch_7[] = { 0x57, 0x3C, 0x0C, 0x00 };
static const BYTE patch_8[] = { 0xF7, 0x3A, 0x0C, 0x00 };
static const BYTE patch_9[] = { 0x63, 0x2A, 0x0C, 0x00 };
static const BYTE patch_10[] = { 0x47, 0x08, 0x0C, 0x00 };
static const BYTE patch_11[] = { 0xE3, 0x06, 0x0C, 0x00 };
static const BYTE patch_12[] = { 0x1F, 0x04, 0x0C, 0x00 };
static const BYTE patch_13[] = { 0xC3, 0x02, 0x0C, 0x00 };
static const BYTE patch_14[] = { 0x8F, 0xFF, 0x0B, 0x00 };
static const BYTE patch_15[] = { 0x1F, 0xF8, 0x0B, 0x00 };
static const BYTE patch_16[] = { 0xC3, 0xF6, 0x0B, 0x00 };
static const BYTE patch_17[] = { 0xAB, 0xF4, 0x0B, 0x00 };
static const BYTE patch_18[] = { 0x39, 0xF4, 0x0B, 0x00 };
static const BYTE patch_19[] = { 0xDD, 0x34 };
static const BYTE patch_20[] = { 0xFF, 0x76 };
static const BYTE patch_21[] = { 0xE8, 0x58, 0x10, 0xED, 0xFF, 0xB9, 0x10, 0x0E, 0x00, 0x00, 0x99, 0xF7, 0xF9, 0x89 };
static const BYTE patch_22[] = { 0xFC, 0xFE };
static const BYTE patch_23[] = { 0xC6 };
static const BYTE patch_24[] = { 0x00, 0xFF };
static const BYTE patch_25[] = { 0x00, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90 };
static const BYTE patch_26[] = { 0xEE, 0x3B, 0x05, 0x00, 0x89, 0x95, 0x0C, 0xFF, 0xFF, 0xFF, 0xC6, 0x85, 0x10, 0xFF, 0xFF };
static const BYTE patch_27[] = { 0x00, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90 };
static const BYTE patch_28[] = { 0x64, 0x25, 0x73, 0x20, 0x25, 0x64 };
static const BYTE patch_29[] = { 0x73, 0x00, 0x00 };
static const BYTE patch_30[] = { 0xE8, 0x8F, 0xF6, 0x04, 0x00 };
static const BYTE patch_31[] = { 0xE8, 0x86, 0x42, 0x00 };
static const BYTE patch_32[] = { 0xE9, 0x04, 0x2F, 0x00, 0x00, 0x90, 0x90 };
static const BYTE patch_33[] = { 0x53, 0x56, 0x8B, 0xDA, 0x8B, 0xF0, 0x66, 0xF7, 0x06, 0xE8, 0xBF, 0x74, 0x07, 0x8B, 0xC6, 0xE8, 0xD9, 0xE4, 0xE1, 0xFF, 0xC7, 0x06, 0x03 };
static const BYTE patch_34[] = { 0xC7, 0x46, 0x04 };
static const BYTE patch_35[] = { 0x33, 0xC0, 0x84, 0xDB, 0x0F, 0x95, 0xC0, 0x89, 0x46, 0x08, 0xC7, 0x46, 0x0C };
static const BYTE patch_36[] = { 0x5E, 0x5B, 0xC3 };
static const BYTE patch_37[] = { 0x55, 0x8B, 0xEC, 0xFF, 0x75, 0x08, 0xE8, 0x12, 0xD4, 0xE7, 0xFF, 0xB9, 0x3C };
static const BYTE patch_38[] = { 0x99, 0xF7, 0xF9, 0x99, 0xF7, 0xF9, 0x8B, 0xE5, 0x5D, 0xC2, 0x04 };
static const BYTE patch_39[] = { 0xE8, 0xA0, 0x17 };
static const BYTE patch_40[] = { 0xE9, 0x2F, 0x1A };
static const BYTE patch_41[] = { 0xE8, 0x30, 0x19 };
static const BYTE patch_42[] = { 0xE8, 0xC3, 0x16 };
static const BYTE patch_43[] = { 0xE8, 0x1E, 0x16 };
static const BYTE patch_44[] = { 0xE8, 0x2D, 0x16 };
static const BYTE patch_45[] = { 0xE8, 0x3C, 0x16 };
static const BYTE patch_46[] = { 0xE8, 0x6F, 0x16 };
static const BYTE patch_47[] = { 0xE8, 0x96, 0x16 };
static const BYTE patch_48[] = { 0xC3, 0x33, 0xC0, 0x89, 0x43, 0x20, 0xC6, 0x43, 0x24 };
static const BYTE patch_49[] = { 0xC3 };
static const BYTE patch_50[] = { 0x80, 0x3D, 0xBC, 0xE5, 0x5F };
static const BYTE patch_51[] = { 0x75, 0x05, 0xE8, 0xD6, 0xE6, 0xE7, 0xFF, 0x8D, 0x45, 0xF4, 0xE8, 0xAE, 0x2A, 0xE1, 0xFF, 0xE9, 0x35, 0x6E, 0xEB, 0xFF };
static const BYTE patch_52[] = { 0xE8, 0xEF, 0xF8, 0xFF, 0xFF, 0xE8, 0x8A, 0xF9, 0xFF, 0xFF, 0xE8, 0xD1, 0xFA, 0xFF, 0xFF, 0xE8, 0xC8, 0xFB, 0xFF, 0xFF, 0xE8, 0xB3, 0xFC, 0xFF, 0xFF, 0xE8, 0x5A, 0xF9, 0xFF, 0xFF, 0xE8, 0x35, 0xFB, 0xFF, 0xFF, 0xE8, 0xC3, 0xE5, 0xFF, 0xFF, 0xA1, 0xC0, 0x9F, 0x5F };
static const BYTE patch_53[] = { 0x8B };
static const BYTE patch_54[] = { 0xE8, 0xAC, 0xF6, 0xED, 0xFF, 0xE8, 0x81, 0xE6, 0xF4, 0xFF, 0xE9, 0x87, 0xB6, 0xFF, 0xFF, 0x90, 0x90 };

static const PatchRegion PATCHES[] = {
    { 0x228, sizeof(patch_0), patch_0 },
    { 0xF242, sizeof(patch_1), patch_1 },
    { 0x7A38F, sizeof(patch_2), patch_2 },
    { 0xAA828, sizeof(patch_3), patch_3 },
    { 0x12F450, sizeof(patch_4), patch_4 },
    { 0x12F5B8, sizeof(patch_5), patch_5 },
    { 0x12FB64, sizeof(patch_6), patch_6 },
    { 0x12FCC8, sizeof(patch_7), patch_7 },
    { 0x12FE28, sizeof(patch_8), patch_8 },
    { 0x130EBC, sizeof(patch_9), patch_9 },
    { 0x1330D8, sizeof(patch_10), patch_10 },
    { 0x13323C, sizeof(patch_11), patch_11 },
    { 0x133500, sizeof(patch_12), patch_12 },
    { 0x13365C, sizeof(patch_13), patch_13 },
    { 0x133990, sizeof(patch_14), patch_14 },
    { 0x134100, sizeof(patch_15), patch_15 },
    { 0x13425C, sizeof(patch_16), patch_16 },
    { 0x134474, sizeof(patch_17), patch_17 },
    { 0x1344E6, sizeof(patch_18), patch_18 },
    { 0x14360F, sizeof(patch_19), patch_19 },
    { 0x19FD28, sizeof(patch_20), patch_20 },
    { 0x19FD2B, sizeof(patch_21), patch_21 },
    { 0x19FD3A, sizeof(patch_22), patch_22 },
    { 0x19FD3E, sizeof(patch_23), patch_23 },
    { 0x19FD40, sizeof(patch_24), patch_24 },
    { 0x19FD44, sizeof(patch_25), patch_25 },
    { 0x19FD79, sizeof(patch_26), patch_26 },
    { 0x19FD89, sizeof(patch_27), patch_27 },
    { 0x1A0191, sizeof(patch_28), patch_28 },
    { 0x1A0198, sizeof(patch_29), patch_29 },
    { 0x1A433B, sizeof(patch_30), patch_30 },
    { 0x1F0A79, sizeof(patch_31), patch_31 },
    { 0x1F0A7E, sizeof(patch_32), patch_32 },
    { 0x1F3923, sizeof(patch_33), patch_33 },
    { 0x1F393D, sizeof(patch_34), patch_34 },
    { 0x1F3944, sizeof(patch_35), patch_35 },
    { 0x1F3955, sizeof(patch_36), patch_36 },
    { 0x1F396B, sizeof(patch_37), patch_37 },
    { 0x1F397B, sizeof(patch_38), patch_38 },
    { 0x1F3987, sizeof(patch_39), patch_39 },
    { 0x1F398C, sizeof(patch_40), patch_40 },
    { 0x1F39AB, sizeof(patch_41), patch_41 },
    { 0x1F39B0, sizeof(patch_42), patch_42 },
    { 0x1F39B5, sizeof(patch_43), patch_43 },
    { 0x1F39BA, sizeof(patch_44), patch_44 },
    { 0x1F39BF, sizeof(patch_45), patch_45 },
    { 0x1F39C4, sizeof(patch_46), patch_46 },
    { 0x1F39C9, sizeof(patch_47), patch_47 },
    { 0x1F39CE, sizeof(patch_48), patch_48 },
    { 0x1F39D8, sizeof(patch_49), patch_49 },
    { 0x1F39E0, sizeof(patch_50), patch_50 },
    { 0x1F39E7, sizeof(patch_51), patch_51 },
    { 0x1F4DC0, sizeof(patch_52), patch_52 },
    { 0x1F4DED, sizeof(patch_53), patch_53 },
    { 0x1F4DEF, sizeof(patch_54), patch_54 },
};

static const size_t PATCH_COUNT = sizeof(PATCHES) / sizeof(PATCHES[0]);

/* ---------------- SHA-256 ---------------- */

typedef struct {
    uint32_t state[8];
    uint64_t bitlen;
    BYTE data[64];
    size_t datalen;
} SHA256_CTX_LOCAL;

static const uint32_t k256[64] = {
    0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,
    0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,
    0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,
    0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,
    0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,
    0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,
    0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,
    0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2
};

#define ROTR(x,n) (((x) >> (n)) | ((x) << (32-(n))))
#define CH(x,y,z) (((x)&(y)) ^ (~(x)&(z)))
#define MAJ(x,y,z) (((x)&(y)) ^ ((x)&(z)) ^ ((y)&(z)))
#define EP0(x) (ROTR((x),2) ^ ROTR((x),13) ^ ROTR((x),22))
#define EP1(x) (ROTR((x),6) ^ ROTR((x),11) ^ ROTR((x),25))
#define SIG0(x) (ROTR((x),7) ^ ROTR((x),18) ^ ((x)>>3))
#define SIG1(x) (ROTR((x),17) ^ ROTR((x),19) ^ ((x)>>10))

static void sha256_transform(SHA256_CTX_LOCAL *ctx, const BYTE data[64]) {
    uint32_t a,b,c,d,e,f,g,h,t1,t2,m[64];
    int i;

    for (i=0;i<16;i++)
        m[i] = ((uint32_t)data[i*4]<<24) | ((uint32_t)data[i*4+1]<<16) |
               ((uint32_t)data[i*4+2]<<8) | (uint32_t)data[i*4+3];
    for (i=16;i<64;i++)
        m[i] = SIG1(m[i-2]) + m[i-7] + SIG0(m[i-15]) + m[i-16];

    a=ctx->state[0]; b=ctx->state[1]; c=ctx->state[2]; d=ctx->state[3];
    e=ctx->state[4]; f=ctx->state[5]; g=ctx->state[6]; h=ctx->state[7];

    for (i=0;i<64;i++) {
        t1 = h + EP1(e) + CH(e,f,g) + k256[i] + m[i];
        t2 = EP0(a) + MAJ(a,b,c);
        h=g; g=f; f=e; e=d+t1;
        d=c; c=b; b=a; a=t1+t2;
    }

    ctx->state[0]+=a; ctx->state[1]+=b; ctx->state[2]+=c; ctx->state[3]+=d;
    ctx->state[4]+=e; ctx->state[5]+=f; ctx->state[6]+=g; ctx->state[7]+=h;
}

static void sha256_init(SHA256_CTX_LOCAL *ctx) {
    ctx->datalen=0; ctx->bitlen=0;
    ctx->state[0]=0x6a09e667; ctx->state[1]=0xbb67ae85;
    ctx->state[2]=0x3c6ef372; ctx->state[3]=0xa54ff53a;
    ctx->state[4]=0x510e527f; ctx->state[5]=0x9b05688c;
    ctx->state[6]=0x1f83d9ab; ctx->state[7]=0x5be0cd19;
}

static void sha256_update(SHA256_CTX_LOCAL *ctx, const BYTE *data, size_t len) {
    size_t i;
    for (i=0;i<len;i++) {
        ctx->data[ctx->datalen++] = data[i];
        if (ctx->datalen == 64) {
            sha256_transform(ctx, ctx->data);
            ctx->bitlen += 512;
            ctx->datalen = 0;
        }
    }
}

static void sha256_final(SHA256_CTX_LOCAL *ctx, BYTE hash[32]) {
    size_t i = ctx->datalen;
    int j;

    ctx->data[i++] = 0x80;
    if (i > 56) {
        while (i < 64) ctx->data[i++] = 0;
        sha256_transform(ctx, ctx->data);
        i = 0;
    }
    while (i < 56) ctx->data[i++] = 0;

    ctx->bitlen += (uint64_t)ctx->datalen * 8;
    for (j=0;j<8;j++)
        ctx->data[63-j] = (BYTE)(ctx->bitlen >> (j*8));

    sha256_transform(ctx, ctx->data);

    for (i=0;i<8;i++) {
        hash[i*4]   = (BYTE)(ctx->state[i] >> 24);
        hash[i*4+1] = (BYTE)(ctx->state[i] >> 16);
        hash[i*4+2] = (BYTE)(ctx->state[i] >> 8);
        hash[i*4+3] = (BYTE)(ctx->state[i]);
    }
}

static BOOL sha256_file(const WCHAR *path, char out[65]) {
    HANDLE h;
    BYTE buf[65536], hash[32];
    DWORD got;
    SHA256_CTX_LOCAL ctx;
    int i;

    h = CreateFileW(path, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING,
                    FILE_ATTRIBUTE_NORMAL, NULL);
    if (h == INVALID_HANDLE_VALUE) return FALSE;

    sha256_init(&ctx);
    while (ReadFile(h, buf, sizeof(buf), &got, NULL) && got)
        sha256_update(&ctx, buf, got);

    if (GetLastError() != ERROR_SUCCESS && GetLastError() != ERROR_HANDLE_EOF) {
        CloseHandle(h);
        return FALSE;
    }
    CloseHandle(h);

    sha256_final(&ctx, hash);
    for (i=0;i<32;i++) sprintf(out+i*2, "%02x", hash[i]);
    out[64] = 0;
    return TRUE;
}

/* ---------------- Helpers ---------------- */

static void info(const WCHAR *text) {
    MessageBoxW(NULL, text, APP_TITLE, MB_OK | MB_ICONINFORMATION);
}

static void warn(const WCHAR *text) {
    MessageBoxW(NULL, text, APP_TITLE, MB_OK | MB_ICONWARNING);
}

static void error_box(const WCHAR *text) {
    MessageBoxW(NULL, text, APP_TITLE, MB_OK | MB_ICONERROR);
}

static BOOL file_exists(const WCHAR *path) {
    DWORD a = GetFileAttributesW(path);
    return a != INVALID_FILE_ATTRIBUTES && !(a & FILE_ATTRIBUTE_DIRECTORY);
}

static BOOL dir_exists(const WCHAR *path) {
    DWORD a = GetFileAttributesW(path);
    return a != INVALID_FILE_ATTRIBUTES && (a & FILE_ATTRIBUTE_DIRECTORY);
}

static void dirname_inplace(WCHAR *path) {
    WCHAR *p = wcsrchr(path, L'\\');
    if (p) *p = 0;
}

static BOOL join_path(WCHAR *out, size_t cap, const WCHAR *a, const WCHAR *b) {
    size_t la = wcslen(a), lb = wcslen(b);
    if (la + 1 + lb + 1 > cap) return FALSE;
    wcscpy(out, a);
    if (la && out[la-1] != L'\\') wcscat(out, L"\\");
    wcscat(out, b);
    return TRUE;
}

static BOOL copy_file_if_missing(const WCHAR *src, const WCHAR *dst) {
    if (file_exists(dst)) return TRUE;
    return CopyFileW(src, dst, TRUE);
}

static int classify_exe(const WCHAR *path) {
    char hash[65];
    if (!sha256_file(path, hash)) return -1;
    if (_stricmp(hash, ORIGINAL_SHA256) == 0) return 1;
    if (_stricmp(hash, PATCHED_SHA256) == 0) return 2;
    return 0;
}

/* Search an uninstall registry branch for an Aquadelic GT install. */
static BOOL search_uninstall_key(HKEY root, const WCHAR *subkey, WCHAR *out, size_t cap, BOOL *unsupported) {
    HKEY hRoot;
    DWORD index = 0;
    WCHAR name[512];

    if (RegOpenKeyExW(root, subkey, 0, KEY_READ, &hRoot) != ERROR_SUCCESS)
        return FALSE;

    for (;;) {
        DWORD nameLen = (DWORD)(sizeof(name)/sizeof(name[0]));
        LONG r = RegEnumKeyExW(hRoot, index++, name, &nameLen, NULL, NULL, NULL, NULL);
        HKEY hApp;
        WCHAR display[512], install[MAX_GAME_PATH], candidate[MAX_GAME_PATH];
        DWORD type, size;
        int cls;

        if (r == ERROR_NO_MORE_ITEMS) break;
        if (r != ERROR_SUCCESS) continue;
        if (RegOpenKeyExW(hRoot, name, 0, KEY_READ, &hApp) != ERROR_SUCCESS) continue;

        size = sizeof(display);
        display[0] = 0;
        if (RegQueryValueExW(hApp, L"DisplayName", NULL, &type, (BYTE*)display, &size) != ERROR_SUCCESS ||
            type != REG_SZ) {
            RegCloseKey(hApp);
            continue;
        }

        if (wcsstr(display, L"Aquadelic") == NULL && wcsstr(display, L"AQUADELIC") == NULL) {
            RegCloseKey(hApp);
            continue;
        }

        size = sizeof(install);
        install[0] = 0;
        if (RegQueryValueExW(hApp, L"InstallLocation", NULL, &type, (BYTE*)install, &size) == ERROR_SUCCESS &&
            (type == REG_SZ || type == REG_EXPAND_SZ) && install[0]) {
            if (join_path(candidate, cap, install, L"Run.exe") && file_exists(candidate)) {
                cls = classify_exe(candidate);
                if (cls == 1 || cls == 2) {
                    wcscpy(out, candidate);
                    RegCloseKey(hApp);
                    RegCloseKey(hRoot);
                    return TRUE;
                }
                *unsupported = TRUE;
                wcscpy(out, candidate);
            }
        }
        RegCloseKey(hApp);
    }

    RegCloseKey(hRoot);
    return FALSE;
}

static BOOL test_candidate(const WCHAR *path, WCHAR *out, size_t cap, BOOL *unsupported) {
    int cls;
    if (!file_exists(path)) return FALSE;
    cls = classify_exe(path);
    if (cls == 1 || cls == 2) {
        wcsncpy(out, path, cap-1);
        out[cap-1] = 0;
        return TRUE;
    }
    if (!*unsupported) {
        *unsupported = TRUE;
        wcsncpy(out, path, cap-1);
        out[cap-1] = 0;
    }
    return FALSE;
}

static BOOL try_base(WCHAR *out, size_t cap, const WCHAR *base, BOOL *unsupported) {
    const WCHAR *rel[] = {
        L"Aquadelic GT\\Run.exe",
        L"AquadelicGT\\Run.exe",
        L"Hammerware\\Aquadelic GT\\Run.exe",
        L"Arcade Moon\\Aquadelic GT\\Run.exe",
        L"Steam\\steamapps\\common\\Aquadelic GT\\Run.exe",
        L"Steam\\SteamApps\\common\\Aquadelic GT\\Run.exe"
    };
    WCHAR p[MAX_GAME_PATH];
    size_t i;

    if (!base || !base[0]) return FALSE;
    for (i=0;i<sizeof(rel)/sizeof(rel[0]);i++) {
        if (join_path(p, MAX_GAME_PATH, base, rel[i]) &&
            test_candidate(p, out, cap, unsupported))
            return TRUE;
    }
    return FALSE;
}

static BOOL auto_find(WCHAR *out, size_t cap, BOOL *unsupported) {
    WCHAR p[MAX_GAME_PATH], dir[MAX_GAME_PATH], env[MAX_GAME_PATH];
    DWORD n;

    *unsupported = FALSE;
    out[0] = 0;

    /* Next to patcher. */
    n = GetModuleFileNameW(NULL, p, MAX_GAME_PATH);
    if (n && n < MAX_GAME_PATH) {
        wcscpy(dir, p);
        dirname_inplace(dir);
        if (join_path(p, MAX_GAME_PATH, dir, L"Run.exe") &&
            test_candidate(p, out, cap, unsupported))
            return TRUE;

        /* One directory above patcher. */
        dirname_inplace(dir);
        if (join_path(p, MAX_GAME_PATH, dir, L"Run.exe") &&
            test_candidate(p, out, cap, unsupported))
            return TRUE;
    }

    /* Current directory. */
    n = GetCurrentDirectoryW(MAX_GAME_PATH, dir);
    if (n && n < MAX_GAME_PATH &&
        join_path(p, MAX_GAME_PATH, dir, L"Run.exe") &&
        test_candidate(p, out, cap, unsupported))
        return TRUE;

    /* Common direct install locations. */
    if (test_candidate(L"C:\\Aquadelic GT\\Run.exe", out, cap, unsupported)) return TRUE;
    if (test_candidate(L"C:\\Games\\Aquadelic GT\\Run.exe", out, cap, unsupported)) return TRUE;
    if (test_candidate(L"C:\\Gry\\Aquadelic GT\\Run.exe", out, cap, unsupported)) return TRUE;

    n = GetEnvironmentVariableW(L"ProgramFiles", env, MAX_GAME_PATH);
    if (n && n < MAX_GAME_PATH && try_base(out, cap, env, unsupported)) return TRUE;

    n = GetEnvironmentVariableW(L"ProgramFiles(x86)", env, MAX_GAME_PATH);
    if (n && n < MAX_GAME_PATH && try_base(out, cap, env, unsupported)) return TRUE;

    /* Installed-program registry, both 32-bit and common 64-bit branch names. */
    if (search_uninstall_key(HKEY_CURRENT_USER,
        L"SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
        out, cap, unsupported)) return TRUE;

    if (search_uninstall_key(HKEY_LOCAL_MACHINE,
        L"SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
        out, cap, unsupported)) return TRUE;

    if (search_uninstall_key(HKEY_LOCAL_MACHINE,
        L"SOFTWARE\\Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall",
        out, cap, unsupported)) return TRUE;

    return FALSE;
}

static BOOL choose_executable(WCHAR *path, DWORD cap) {
    OPENFILENAMEW ofn;
    WCHAR filter[] =
        L"Executable files (*.exe)\0*.exe\0"
        L"All files (*.*)\0*.*\0\0";

    ZeroMemory(&ofn, sizeof(ofn));
    path[0] = 0;

    ofn.lStructSize = sizeof(ofn);
    ofn.hwndOwner = NULL;
    ofn.lpstrFilter = filter;
    ofn.nFilterIndex = 1;
    ofn.lpstrFile = path;
    ofn.nMaxFile = cap;
    ofn.lpstrTitle = L"Select your original Aquadelic GT executable";
    ofn.Flags = OFN_EXPLORER | OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST;

    return GetOpenFileNameW(&ofn);
}

static BOOL patch_file(const WCHAR *path) {
    HANDLE h;
    LARGE_INTEGER size;
    BYTE *data = NULL;
    DWORD got, written;
    size_t i;
    WCHAR gameDir[MAX_GAME_PATH], backupDir[MAX_GAME_PATH], backupPath[MAX_GAME_PATH];
    WCHAR tempPath[MAX_GAME_PATH], msg[1024];
    char hash[65];

    h = CreateFileW(path, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING,
                    FILE_ATTRIBUTE_NORMAL, NULL);
    if (h == INVALID_HANDLE_VALUE) {
        error_box(L"Could not read the selected executable.\nNo files were changed.");
        return FALSE;
    }

    if (!GetFileSizeEx(h, &size) || size.QuadPart <= 0 || size.QuadPart > 0x7fffffff) {
        CloseHandle(h);
        error_box(L"Could not read the selected executable.\nNo files were changed.");
        return FALSE;
    }

    data = (BYTE*)HeapAlloc(GetProcessHeap(), 0, (SIZE_T)size.QuadPart);
    if (!data) {
        CloseHandle(h);
        error_box(L"Could not allocate memory.\nNo files were changed.");
        return FALSE;
    }

    if (!ReadFile(h, data, (DWORD)size.QuadPart, &got, NULL) || got != (DWORD)size.QuadPart) {
        CloseHandle(h);
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"Could not read the selected executable.\nNo files were changed.");
        return FALSE;
    }
    CloseHandle(h);

    for (i=0;i<PATCH_COUNT;i++) {
        if ((uint64_t)PATCHES[i].offset + PATCHES[i].size > (uint64_t)size.QuadPart) {
            HeapFree(GetProcessHeap(), 0, data);
            error_box(L"Internal patch verification failed.\nNo files were changed.");
            return FALSE;
        }
        memcpy(data + PATCHES[i].offset, PATCHES[i].data, PATCHES[i].size);
    }

    /* Prepare paths. */
    wcsncpy(gameDir, path, MAX_GAME_PATH-1);
    gameDir[MAX_GAME_PATH-1] = 0;
    dirname_inplace(gameDir);

    if (!join_path(backupDir, MAX_GAME_PATH, gameDir, BACKUP_DIR) ||
        !join_path(backupPath, MAX_GAME_PATH, backupDir, L"Run.exe")) {
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"The selected path is too long.\nNo files were changed.");
        return FALSE;
    }

    if (!dir_exists(backupDir) && !CreateDirectoryW(backupDir, NULL) &&
        GetLastError() != ERROR_ALREADY_EXISTS) {
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"Could not create the backup folder.\nNo files were changed.\n"
                  L"If the game is installed in Program Files, try running the patcher as Administrator.");
        return FALSE;
    }

    if (!copy_file_if_missing(path, backupPath)) {
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"Could not back up the original executable.\nNo files were changed.\n"
                  L"If the game is installed in Program Files, try running the patcher as Administrator.");
        return FALSE;
    }

    if (wcslen(path) + 20 >= MAX_GAME_PATH) {
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"The selected path is too long.\nNo files were changed.");
        return FALSE;
    }
    wcscpy(tempPath, path);
    wcscat(tempPath, L".aquadelic_patch.tmp");

    h = CreateFileW(tempPath, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS,
                    FILE_ATTRIBUTE_NORMAL, NULL);
    if (h == INVALID_HANDLE_VALUE) {
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"Could not write the patched executable.\nThe original executable was not replaced.\n"
                  L"If the game is installed in Program Files, try running the patcher as Administrator.");
        return FALSE;
    }

    if (!WriteFile(h, data, (DWORD)size.QuadPart, &written, NULL) ||
        written != (DWORD)size.QuadPart) {
        CloseHandle(h);
        DeleteFileW(tempPath);
        HeapFree(GetProcessHeap(), 0, data);
        error_box(L"Could not write the patched executable.\nThe original executable was not replaced.");
        return FALSE;
    }
    FlushFileBuffers(h);
    CloseHandle(h);
    HeapFree(GetProcessHeap(), 0, data);

    if (!sha256_file(tempPath, hash) || _stricmp(hash, PATCHED_SHA256) != 0) {
        DeleteFileW(tempPath);
        error_box(L"Patched executable verification failed.\nThe original executable was not replaced.");
        return FALSE;
    }

    if (!MoveFileExW(tempPath, path, MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH)) {
        DeleteFileW(tempPath);
        error_box(L"Could not replace the executable.\nMake sure the game is closed and try running the patcher as Administrator.");
        return FALSE;
    }

    if (!sha256_file(path, hash) || _stricmp(hash, PATCHED_SHA256) != 0) {
        error_box(L"The executable was replaced, but final verification could not be completed.");
        return FALSE;
    }

    wsprintfW(msg,
        L"Patch installed successfully.\n\n"
        L"The original executable was saved in:\n"
        L"%s\\Run.exe",
        BACKUP_DIR);
    info(msg);
    return TRUE;
}

int WINAPI WinMain(HINSTANCE hInst, HINSTANCE hPrev, LPSTR cmdLine, int show) {
    WCHAR path[MAX_GAME_PATH];
    BOOL unsupported = FALSE;
    int cls;

    (void)hInst; (void)hPrev; (void)cmdLine; (void)show;

    if (!auto_find(path, MAX_GAME_PATH, &unsupported)) {
        if (unsupported && path[0]) {
            warn(L"This executable is not compatible with this patcher.\n\n"
                 L"It may be a different release, language edition, Steam version, or another game revision.\n\n"
                 L"No files were changed.\n\n"
                 L"Please contact the creator of the patcher and provide your original game executable so support for this version can be investigated.");
            return 0;
        }

        if (!choose_executable(path, MAX_GAME_PATH))
            return 0;
    }

    cls = classify_exe(path);

    if (cls < 0) {
        error_box(L"Could not read the selected executable.\nNo files were changed.");
        return 0;
    }

    if (cls == 2) {
        info(L"This executable is already patched.\n\nNo files were changed.");
        return 0;
    }

    if (cls == 0) {
        warn(L"This executable is not compatible with this patcher.\n\n"
             L"It may be a different release, language edition, Steam version, or another game revision.\n\n"
             L"No files were changed.\n\n"
             L"Please contact the creator of the patcher and provide your original game executable so support for this version can be investigated.");
        return 0;
    }

    patch_file(path);
    return 0;
}
