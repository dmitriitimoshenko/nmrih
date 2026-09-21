/**
 * NMRiH HUD Bars
 *
 * Draws compact health and stamina bars in a corner of the screen, with the
 * current value printed inside the bar, plus icons for bleeding and infection.
 *
 * Everything is server side: the bars are game_text HUD channels, so the bar
 * itself is built out of block characters. The plugin takes 3 of the 6 HUD
 * channels the engine offers (health, stamina, status).
 */

#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdktools>

#define PLUGIN_VERSION "1.1.0"

/* Bumped whenever the generated config's defaults change. AutoExecConfig never
 * rewrites an existing file, so without this a new look would be silently
 * overridden by the config written for the previous one. */
#define CONFIG_VERSION "2"
#define CONFIG_FILE    "../../cfg/sourcemod/nmrih_hudbars.cfg"
#define CONFIG_STAMP   "../../cfg/sourcemod/nmrih_hudbars.version"

#define MIN_CELLS 4
#define MAX_CELLS 40

public Plugin myinfo =
{
    name        = "NMRiH HUD Bars",
    author      = "dmitriitimoshenko",
    description = "Health/stamina bars with bleeding and infection indicators",
    version     = PLUGIN_VERSION,
    url         = "https://github.com/dmitriitimoshenko/nmrih"
};

Handle g_hHudHealth;
Handle g_hHudStamina;
Handle g_hHudStatus;

ConVar g_cvEnabled;
ConVar g_cvX;
ConVar g_cvY;
ConVar g_cvLineHeight;
ConVar g_cvInterval;
ConVar g_cvCells;
ConVar g_cvStyle;
ConVar g_cvGlyphs;
ConVar g_cvAlpha;
ConVar g_cvNumberInside;
ConVar g_cvLabels;
ConVar g_cvHealthMax;
ConVar g_cvStaminaMax;
ConVar g_cvStaminaColor;
ConVar g_cvIconBleeding;
ConVar g_cvIconInfected;
ConVar g_cvPropStamina;
ConVar g_cvPropBleeding;
ConVar g_cvPropInfected;
ConVar g_cvDetected;

bool g_bHidden[MAXPLAYERS + 1];
bool g_bHudMsgBroken;
Handle g_hTimer;

/* Netprops resolved against a live player, see ResolveProps. */
#define PROP_STAMINA  0
#define PROP_BLEEDING 1
#define PROP_INFECTED 2
#define PROP_COUNT    3

char g_sResolved[PROP_COUNT][64];
bool g_bResolved;
float g_flStaminaSeen;

/* Candidates used by sm_hudbars_scan to find the right netprops on this build. */
char g_sStaminaCandidates[][] = {
    "m_flStamina", "m_flStaminaPercent", "m_flSprintStamina", "m_flPlayerStamina"
};
char g_sBleedingCandidates[][] = {
    "m_bIsBleeding", "m_bBleeding", "m_iBleedingCount", "m_flBleedOutTime", "m_bIsBleedingOut"
};
char g_sInfectedCandidates[][] = {
    "m_bIsInfected", "m_bInfected", "m_bPlayerInfected", "m_iInfectionStage",
    "m_flInfectionTime", "m_flInfectionDeathTime"
};

public void OnPluginStart()
{
    CreateConVar("sm_hudbars_version", PLUGIN_VERSION, "NMRiH HUD Bars version",
        FCVAR_NOTIFY | FCVAR_DONTRECORD);

    g_cvEnabled      = CreateConVar("sm_hudbars_enabled", "1", "Enable the HUD bars", _, true, 0.0, true, 1.0);
    g_cvX            = CreateConVar("sm_hudbars_x", "0.015", "Horizontal position, 0.0 = left, 1.0 = right, -1 = centered", _, true, -1.0, true, 1.0);
    g_cvY            = CreateConVar("sm_hudbars_y", "0.045", "Vertical position of the first line, 0.0 = top, 1.0 = bottom", _, true, -1.0, true, 1.0);
    g_cvLineHeight   = CreateConVar("sm_hudbars_line_height", "0.05", "Distance between the lines", _, true, 0.01, true, 0.2);
    g_cvInterval     = CreateConVar("sm_hudbars_interval", "0.2", "Refresh interval in seconds", _, true, 0.1, true, 1.0);
    g_cvCells        = CreateConVar("sm_hudbars_cells", "10", "Width of a bar in characters", _, true, float(MIN_CELLS), true, float(MAX_CELLS));
    g_cvStyle        = CreateConVar("sm_hudbars_style", "shaded", "Bar look: shaded, blocks, squares, dots, ascii. Applies on the next refresh");
    g_cvGlyphs       = CreateConVar("sm_hudbars_glyphs", "", "Overrides the style with two characters, \"<filled> <empty>\"");
    g_cvAlpha        = CreateConVar("sm_hudbars_alpha", "220", "Opacity of the bars, 0-255", _, true, 0.0, true, 255.0);
    g_cvNumberInside = CreateConVar("sm_hudbars_number_inside", "0", "1 = print the value inside the bar, 0 = after it", _, true, 0.0, true, 1.0);
    g_cvLabels       = CreateConVar("sm_hudbars_labels", "1", "1 = HP/SP before the bars, 0 = no labels, which is what makes the bars line up exactly", _, true, 0.0, true, 1.0);
    g_cvHealthMax    = CreateConVar("sm_hudbars_health_max", "100", "Health value that fills the bar completely", _, true, 1.0);
    g_cvStaminaMax   = CreateConVar("sm_hudbars_stamina_max", "0", "Stamina that fills the bar. 0 = learn it from the highest value seen", _, true, 0.0);
    g_cvStaminaColor = CreateConVar("sm_hudbars_stamina_color", "120 175 220", "Colour of the stamina bar, \"R G B\"");
    g_cvIconBleeding = CreateConVar("sm_hudbars_icon_bleeding", "♦", "Icon shown while bleeding");
    g_cvIconInfected = CreateConVar("sm_hudbars_icon_infected", "▲", "Icon shown while infected");

    /* Netprop names differ between NMRiH builds, so the plugin finds them
     * itself against a live player. These only exist to override that:
     * empty = detect, a name = force it, "none" = hide that element. */
    g_cvPropStamina  = CreateConVar("sm_hudbars_prop_stamina", "", "Stamina netprop. Empty = detect automatically, \"none\" = hide the bar");
    g_cvPropBleeding = CreateConVar("sm_hudbars_prop_bleeding", "", "Bleeding netprop. Empty = detect automatically, \"none\" = hide the icon");
    g_cvPropInfected = CreateConVar("sm_hudbars_prop_infected", "", "Infection netprop. Empty = detect automatically, \"none\" = hide the icon");

    /* Reported rather than configured: FCVAR_NOTIFY puts it in the server's
     * rules, so what the plugin detected can be read without server access. */
    g_cvDetected = CreateConVar("sm_hudbars_detected", "pending",
        "What the netprop detection settled on. Diagnostics only, setting it does nothing",
        FCVAR_NOTIFY | FCVAR_DONTRECORD);

    RegConsoleCmd("sm_hud", Cmd_ToggleHud, "Toggle the health/stamina bars for yourself");
    RegAdminCmd("sm_hudbars_scan", Cmd_Scan, ADMFLAG_GENERIC, "List the netprops this build actually exposes");

    g_hHudHealth  = CreateHudSynchronizer();
    g_hHudStamina = CreateHudSynchronizer();
    g_hHudStatus  = CreateHudSynchronizer();

    g_cvInterval.AddChangeHook(OnIntervalChanged);

    RefreshConfigIfStale();
    AutoExecConfig(true, "nmrih_hudbars");

    /* OnConfigsExecuted does not fire when the plugin is loaded mid-map. */
    RestartTimer();
}

/**
 * Drops the generated config when it predates the current defaults, so a new
 * look actually reaches a server that already ran an older build. Only the
 * plugin's own file is touched, and only when CONFIG_VERSION moves.
 */
void RefreshConfigIfStale()
{
    char stampPath[PLATFORM_MAX_PATH];
    char configPath[PLATFORM_MAX_PATH];
    BuildPath(Path_SM, stampPath, sizeof(stampPath), CONFIG_STAMP);
    BuildPath(Path_SM, configPath, sizeof(configPath), CONFIG_FILE);

    char stamped[16];
    File stamp = OpenFile(stampPath, "r");
    if (stamp != null)
    {
        if (!stamp.ReadLine(stamped, sizeof(stamped)))
        {
            stamped[0] = '\0';
        }
        delete stamp;
        TrimString(stamped);
    }

    if (StrEqual(stamped, CONFIG_VERSION))
    {
        return;
    }

    if (FileExists(configPath) && DeleteFile(configPath))
    {
        LogMessage("defaults changed, regenerating cfg/sourcemod/nmrih_hudbars.cfg");
    }

    stamp = OpenFile(stampPath, "w");
    if (stamp != null)
    {
        stamp.WriteLine("%s", CONFIG_VERSION);
        delete stamp;
    }
}

public void OnConfigsExecuted()
{
    RestartTimer();
}

void OnIntervalChanged(ConVar convar, const char[] oldValue, const char[] newValue)
{
    RestartTimer();
}

void RestartTimer()
{
    delete g_hTimer;
    /* No TIMER_FLAG_NO_MAPCHANGE on purpose: the engine would free the timer on
     * a map change and leave a stale handle behind for the next delete. */
    g_hTimer = CreateTimer(g_cvInterval.FloatValue, Timer_Draw, _, TIMER_REPEAT);
}

public void OnMapStart()
{
    g_bHudMsgBroken = false;
    /* A game update can rename or move a netprop, so never trust a resolution
     * from before the map change. */
    g_bResolved = false;
    g_flStaminaSeen = 0.0;
}

public void OnClientPutInServer(int client)
{
    g_bHidden[client] = false;
}

public Action Timer_Draw(Handle timer)
{
    if (!g_cvEnabled.BoolValue)
    {
        return Plugin_Continue;
    }

    for (int client = 1; client <= MaxClients; client++)
    {
        if (!IsClientInGame(client) || IsFakeClient(client) || g_bHidden[client])
        {
            continue;
        }
        if (!IsPlayerAlive(client))
        {
            continue;
        }
        DrawFor(client);
    }

    return Plugin_Continue;
}

/* Glyph pairs that have been eyeballed in game. "shaded" is the default because
 * a solid block fills the whole line box and reads as a censor bar. */
void StyleGlyphs(char[] full, int fullLen, char[] empty, int emptyLen)
{
    char override[32];
    char parts[2][16];
    g_cvGlyphs.GetString(override, sizeof(override));
    if (ExplodeString(override, " ", parts, sizeof(parts), sizeof(parts[])) == 2)
    {
        strcopy(full, fullLen, parts[0]);
        strcopy(empty, emptyLen, parts[1]);
        return;
    }

    char style[16];
    g_cvStyle.GetString(style, sizeof(style));

    if (StrEqual(style, "blocks", false))
    {
        strcopy(full, fullLen, "█");
        strcopy(empty, emptyLen, "░");
    }
    else if (StrEqual(style, "squares", false))
    {
        strcopy(full, fullLen, "■");
        strcopy(empty, emptyLen, "□");
    }
    else if (StrEqual(style, "dots", false))
    {
        strcopy(full, fullLen, "●");
        strcopy(empty, emptyLen, "○");
    }
    else if (StrEqual(style, "ascii", false))
    {
        strcopy(full, fullLen, "#");
        strcopy(empty, emptyLen, "-");
    }
    else
    {
        strcopy(full, fullLen, "▓");
        strcopy(empty, emptyLen, "░");
    }
}

void DrawFor(int client)
{
    ResolveProps(client);

    char full[16];
    char empty[16];
    StyleGlyphs(full, sizeof(full), empty, sizeof(empty));

    int cells = g_cvCells.IntValue;
    if (cells < MIN_CELLS) cells = MIN_CELLS;
    if (cells > MAX_CELLS) cells = MAX_CELLS;

    bool inside   = g_cvNumberInside.BoolValue;
    float x       = g_cvX.FloatValue;
    float y       = g_cvY.FloatValue;
    float line    = g_cvLineHeight.FloatValue;
    float hold    = g_cvInterval.FloatValue + 0.2;

    int health = GetClientHealth(client);
    float healthFraction = Fraction(float(health), g_cvHealthMax.FloatValue);

    char healthBar[256];
    BuildBar(healthBar, sizeof(healthBar), healthFraction, cells, full, empty, health, inside);

    int r, g, b;
    HealthColour(healthFraction, r, g, b);

    bool labelled = g_cvLabels.BoolValue;

    char healthLine[320];
    ComposeLine(healthLine, sizeof(healthLine), "HP ", healthBar, health, inside, labelled);

    SetHudTextParams(x, y, hold, r, g, b, g_cvAlpha.IntValue, 0, 0.0, 0.0, 0.0);
    if (ShowSyncHudText(client, g_hHudHealth, "%s", healthLine) < 0)
    {
        /* The game has no usable HudMsg user message: fall back to hint text,
         * which cannot be positioned or coloured but at least shows up. */
        DrawFallback(client, healthBar, cells, full, empty, inside);
        return;
    }

    float stamina = ReadProp(client, PROP_STAMINA);
    if (stamina >= 0.0)
    {
        float staminaFraction = Fraction(stamina, StaminaMax(stamina));

        char staminaBar[256];
        BuildBar(staminaBar, sizeof(staminaBar), staminaFraction, cells, full, empty,
            RoundToNearest(stamina), inside);

        char staminaLine[320];
        ComposeLine(staminaLine, sizeof(staminaLine), "SP ", staminaBar,
            RoundToNearest(stamina), inside, labelled);

        ParseColour(g_cvStaminaColor, r, g, b);
        SetHudTextParams(x, y + line, hold, r, g, b, g_cvAlpha.IntValue, 0, 0.0, 0.0, 0.0);
        ShowSyncHudText(client, g_hHudStamina, "%s", staminaLine);
    }

    DrawStatus(client, x, y + line * 2.0, hold);
}

void DrawStatus(int client, float x, float y, float hold)
{
    bool bleeding = ReadFlag(client, PROP_BLEEDING);
    bool infected = ReadFlag(client, PROP_INFECTED);

    if (!bleeding && !infected)
    {
        return;
    }

    char icon[16];
    char status[128];
    status[0] = '\0';

    if (bleeding)
    {
        g_cvIconBleeding.GetString(icon, sizeof(icon));
        Format(status, sizeof(status), "%s BLEEDING", icon);
    }
    if (infected)
    {
        g_cvIconInfected.GetString(icon, sizeof(icon));
        if (status[0] != '\0')
        {
            StrCat(status, sizeof(status), "   ");
        }
        Format(status, sizeof(status), "%s%s INFECTED", status, icon);
    }

    /* Bleeding is the more urgent one, so it wins the single status channel. */
    bool bright = (RoundToFloor(GetGameTime() * 2.5) % 2) == 0;
    int r, g, b;
    if (bleeding)
    {
        r = 235; g = bright ? 60 : 25; b = bright ? 60 : 25;
    }
    else
    {
        r = bright ? 150 : 110; g = 220; b = bright ? 90 : 60;
    }

    SetHudTextParams(x, y, hold, r, g, b, g_cvAlpha.IntValue, 0, 0.0, 0.0, 0.0);
    ShowSyncHudText(client, g_hHudStatus, "%s", status);
}

void DrawFallback(int client, const char[] healthBar, int cells, const char[] full,
    const char[] empty, bool inside)
{
    if (!g_bHudMsgBroken)
    {
        g_bHudMsgBroken = true;
        LogError("ShowSyncHudText failed - this build has no usable HudMsg user message, falling back to hint text.");
    }

    bool labelled = g_cvLabels.BoolValue;

    char line[320];
    ComposeLine(line, sizeof(line), "HP ", healthBar, GetClientHealth(client), inside, labelled);

    float stamina = ReadProp(client, PROP_STAMINA);
    if (stamina >= 0.0)
    {
        char staminaBar[256];
        BuildBar(staminaBar, sizeof(staminaBar), Fraction(stamina, StaminaMax(stamina)),
            cells, full, empty, RoundToNearest(stamina), inside);

        char staminaLine[320];
        ComposeLine(staminaLine, sizeof(staminaLine), "SP ", staminaBar,
            RoundToNearest(stamina), inside, labelled);
        Format(line, sizeof(line), "%s\n%s", line, staminaLine);
    }

    char icon[16];
    if (ReadFlag(client, PROP_BLEEDING))
    {
        g_cvIconBleeding.GetString(icon, sizeof(icon));
        Format(line, sizeof(line), "%s\n%s BLEEDING", line, icon);
    }
    if (ReadFlag(client, PROP_INFECTED))
    {
        g_cvIconInfected.GetString(icon, sizeof(icon));
        Format(line, sizeof(line), "%s\n%s INFECTED", line, icon);
    }

    PrintHintText(client, "%s", line);
}

/**
 * Builds "███72███░░░": cells of full/empty characters with the value written
 * over the middle of them. The cells are kept as separate strings on purpose,
 * the block characters are multi-byte and cannot be indexed as bytes.
 */
void BuildBar(char[] out, int maxlen, float fraction, int cells, const char[] full,
    const char[] empty, int value, bool inside)
{
    char number[16];
    IntToString(value, number, sizeof(number));
    int numberLength = strlen(number);

    int filled = RoundToNearest(fraction * float(cells));
    if (filled < 0) filled = 0;
    if (filled > cells) filled = cells;

    int numberStart = -1;
    if (inside && numberLength <= cells)
    {
        numberStart = (cells - numberLength) / 2;
    }

    strcopy(out, maxlen, "");
    for (int i = 0; i < cells; i++)
    {
        if (numberStart != -1 && i >= numberStart && i < numberStart + numberLength)
        {
            char digit[2];
            digit[0] = number[i - numberStart];
            digit[1] = '\0';
            StrCat(out, maxlen, digit);
        }
        else
        {
            StrCat(out, maxlen, (i < filled) ? full : empty);
        }
    }

}

/**
 * The label sits before the bar, which means the bars start wherever the label
 * ends - and the HUD font is proportional, so "HP" and "SP" are not the same
 * width and the two bars do not line up exactly. Turning labels off is what
 * makes them line up: every line then starts with the same bracket.
 */
void ComposeLine(char[] out, int maxlen, const char[] label, const char[] bar,
    int value, bool inside, bool labelled)
{
    if (inside)
    {
        Format(out, maxlen, "%s[%s]", labelled ? label : "", bar);
        return;
    }
    Format(out, maxlen, "%s[%s] %d", labelled ? label : "", bar, value);
}

/* Green when healthy, through yellow, to red when nearly dead. */
void HealthColour(float fraction, int &r, int &g, int &b)
{
    if (fraction <= 0.5)
    {
        float t = fraction / 0.5;
        r = 215;
        g = RoundToNearest(70.0 + t * 130.0);
        b = RoundToNearest(65.0 + t * 25.0);
    }
    else
    {
        float t = (fraction - 0.5) / 0.5;
        r = RoundToNearest(215.0 - t * 90.0);
        g = RoundToNearest(200.0 - t * 5.0);
        b = RoundToNearest(90.0 + t * 25.0);
    }
}

void ParseColour(ConVar convar, int &r, int &g, int &b)
{
    char value[32];
    char parts[3][8];
    convar.GetString(value, sizeof(value));

    if (ExplodeString(value, " ", parts, sizeof(parts), sizeof(parts[])) == 3)
    {
        r = StringToInt(parts[0]);
        g = StringToInt(parts[1]);
        b = StringToInt(parts[2]);
        return;
    }

    r = 90; g = 190; b = 255;
}

/**
 * NMRiH does not agree with everyone about what full stamina is - this build
 * reports 130 - and there is no netprop saying so. Rather than shipping a
 * number that is wrong somewhere else, the highest value ever seen is taken as
 * the maximum. It is right within seconds of play and needs no configuration.
 */
float StaminaMax(float current)
{
    float configured = g_cvStaminaMax.FloatValue;
    if (configured > 0.0)
    {
        return configured;
    }

    if (current > g_flStaminaSeen)
    {
        g_flStaminaSeen = current;
    }
    return (g_flStaminaSeen > 0.0) ? g_flStaminaSeen : 1.0;
}

float Fraction(float value, float max)
{
    if (max <= 0.0)
    {
        return 0.0;
    }
    float fraction = value / max;
    if (fraction < 0.0) return 0.0;
    if (fraction > 1.0) return 1.0;
    return fraction;
}

/**
 * Works out which netprops this build actually has, once, against a player that
 * is in the game. The cvars are only consulted as an override: a name that does
 * not exist here is ignored rather than obeyed, so a config written for another
 * build cannot switch the bars off.
 */
void ResolveProps(int client)
{
    if (g_bResolved)
    {
        return;
    }
    g_bResolved = true;

    ResolveOne(client, PROP_STAMINA, g_cvPropStamina,
        g_sStaminaCandidates, sizeof(g_sStaminaCandidates));
    ResolveOne(client, PROP_BLEEDING, g_cvPropBleeding,
        g_sBleedingCandidates, sizeof(g_sBleedingCandidates));
    ResolveOne(client, PROP_INFECTED, g_cvPropInfected,
        g_sInfectedCandidates, sizeof(g_sInfectedCandidates));

    char stamina[64];
    char bleeding[64];
    char infection[64];
    Describe(PROP_STAMINA, stamina, sizeof(stamina));
    Describe(PROP_BLEEDING, bleeding, sizeof(bleeding));
    Describe(PROP_INFECTED, infection, sizeof(infection));

    char summary[192];
    Format(summary, sizeof(summary), "stamina=%s bleeding=%s infection=%s",
        stamina, bleeding, infection);

    g_cvDetected.SetString(summary);
    LogMessage("netprops: %s", summary);
}

void ResolveOne(int client, int slot, ConVar convar, const char[][] candidates, int count)
{
    g_sResolved[slot][0] = '\0';

    char configured[64];
    convar.GetString(configured, sizeof(configured));

    if (strcmp(configured, "none", false) == 0)
    {
        return;
    }
    if (configured[0] != '\0' && HasEntProp(client, Prop_Send, configured))
    {
        strcopy(g_sResolved[slot], sizeof(g_sResolved[]), configured);
        return;
    }
    if (configured[0] != '\0')
    {
        LogMessage("configured netprop \"%s\" does not exist on this build, detecting instead", configured);
    }

    for (int i = 0; i < count; i++)
    {
        if (HasEntProp(client, Prop_Send, candidates[i]))
        {
            strcopy(g_sResolved[slot], sizeof(g_sResolved[]), candidates[i]);
            return;
        }
    }
}

void Describe(int slot, char[] out, int maxlen)
{
    strcopy(out, maxlen, g_sResolved[slot][0] == '\0' ? "none" : g_sResolved[slot]);
}

/* Reads a resolved netprop as a float, whatever its actual type is.
 * -1.0 = this build does not have it. */
float ReadProp(int client, int slot)
{
    char prop[64];
    strcopy(prop, sizeof(prop), g_sResolved[slot]);
    if (prop[0] == '\0' || !HasEntProp(client, Prop_Send, prop))
    {
        return -1.0;
    }

    char netclass[64];
    if (!GetEntityNetClass(client, netclass, sizeof(netclass)))
    {
        return -1.0;
    }

    PropFieldType type;
    if (FindSendPropInfo(netclass, prop, type) == -1)
    {
        return -1.0;
    }

    if (type == PropField_Float)
    {
        return GetEntPropFloat(client, Prop_Send, prop);
    }
    return float(GetEntProp(client, Prop_Send, prop));
}

/* Anything above zero counts as set, so timer-style props work as flags too. */
bool ReadFlag(int client, int slot)
{
    return ReadProp(client, slot) > 0.0;
}

public Action Cmd_ToggleHud(int client, int args)
{
    if (client == 0)
    {
        ReplyToCommand(client, "[SM] This command is in-game only.");
        return Plugin_Handled;
    }

    g_bHidden[client] = !g_bHidden[client];
    if (g_bHidden[client])
    {
        ClearSyncHud(client, g_hHudHealth);
        ClearSyncHud(client, g_hHudStamina);
        ClearSyncHud(client, g_hHudStatus);
    }

    ReplyToCommand(client, "[SM] HUD bars %s.", g_bHidden[client] ? "hidden" : "shown");
    return Plugin_Handled;
}

/**
 * Prints which of the candidate netprops this build actually has, together with
 * their current value, so the sm_hudbars_prop_* cvars can be pointed at the
 * right ones without recompiling anything.
 */
public Action Cmd_Scan(int client, int args)
{
    /* Falls back to any connected player so this works from the server console
     * and over rcon, where there is no caller to read props off. */
    int target = (client > 0 && IsClientInGame(client)) ? client : FirstPlayer();
    if (target == 0)
    {
        ReplyToCommand(client, "[SM] Nobody is connected: the props are read off a live player.");
        return Plugin_Handled;
    }

    ResolveProps(target);

    char netclass[64];
    GetEntityNetClass(target, netclass, sizeof(netclass));
    ReplyToCommand(client, "[SM] Network class: %s", netclass);
    char stamina[64];
    char bleeding[64];
    char infection[64];
    Describe(PROP_STAMINA, stamina, sizeof(stamina));
    Describe(PROP_BLEEDING, bleeding, sizeof(bleeding));
    Describe(PROP_INFECTED, infection, sizeof(infection));
    ReplyToCommand(client, "[SM] In use: stamina=%s bleeding=%s infection=%s",
        stamina, bleeding, infection);

    ScanGroup(target, client, netclass, "stamina", g_sStaminaCandidates, sizeof(g_sStaminaCandidates));
    ScanGroup(target, client, netclass, "bleeding", g_sBleedingCandidates, sizeof(g_sBleedingCandidates));
    ScanGroup(target, client, netclass, "infection", g_sInfectedCandidates, sizeof(g_sInfectedCandidates));

    ReplyToCommand(client, "[SM] Nothing useful above? Run sm_dump_netprops_xml props.xml and search it for the class shown here.");
    return Plugin_Handled;
}

int FirstPlayer()
{
    for (int client = 1; client <= MaxClients; client++)
    {
        if (IsClientInGame(client) && !IsFakeClient(client))
        {
            return client;
        }
    }
    return 0;
}

void ScanGroup(int target, int reply, const char[] netclass, const char[] label,
    const char[][] candidates, int count)
{
    ReplyToCommand(reply, "[SM] --- %s ---", label);

    bool found = false;
    for (int i = 0; i < count; i++)
    {
        if (!HasEntProp(target, Prop_Send, candidates[i]))
        {
            continue;
        }

        found = true;

        PropFieldType type;
        if (FindSendPropInfo(netclass, candidates[i], type) != -1 && type == PropField_Float)
        {
            ReplyToCommand(reply, "[SM]   %s = %.2f (float)", candidates[i],
                GetEntPropFloat(target, Prop_Send, candidates[i]));
        }
        else
        {
            ReplyToCommand(reply, "[SM]   %s = %d (int)", candidates[i],
                GetEntProp(target, Prop_Send, candidates[i]));
        }
    }

    if (!found)
    {
        ReplyToCommand(reply, "[SM]   none of the known names exist here");
    }
}
