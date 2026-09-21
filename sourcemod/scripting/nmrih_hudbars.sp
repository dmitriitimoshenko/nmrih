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

#define PLUGIN_VERSION "1.0.0"

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
ConVar g_cvCharFull;
ConVar g_cvCharEmpty;
ConVar g_cvNumberInside;
ConVar g_cvHealthMax;
ConVar g_cvStaminaMax;
ConVar g_cvStaminaColor;
ConVar g_cvIconBleeding;
ConVar g_cvIconInfected;
ConVar g_cvPropStamina;
ConVar g_cvPropBleeding;
ConVar g_cvPropInfected;

bool g_bHidden[MAXPLAYERS + 1];
bool g_bHudMsgBroken;
Handle g_hTimer;

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
    g_cvLineHeight   = CreateConVar("sm_hudbars_line_height", "0.035", "Distance between the lines", _, true, 0.01, true, 0.2);
    g_cvInterval     = CreateConVar("sm_hudbars_interval", "0.2", "Refresh interval in seconds", _, true, 0.1, true, 1.0);
    g_cvCells        = CreateConVar("sm_hudbars_cells", "12", "Width of a bar in characters", _, true, float(MIN_CELLS), true, float(MAX_CELLS));
    g_cvCharFull     = CreateConVar("sm_hudbars_char_full", "█", "Character of a filled cell. Use # if the font has no block glyphs");
    g_cvCharEmpty    = CreateConVar("sm_hudbars_char_empty", "░", "Character of an empty cell. Use - if the font has no block glyphs");
    g_cvNumberInside = CreateConVar("sm_hudbars_number_inside", "1", "1 = print the value inside the bar, 0 = after it", _, true, 0.0, true, 1.0);
    g_cvHealthMax    = CreateConVar("sm_hudbars_health_max", "100", "Health value that fills the bar completely", _, true, 1.0);
    g_cvStaminaMax   = CreateConVar("sm_hudbars_stamina_max", "100", "Stamina value that fills the bar completely", _, true, 1.0);
    g_cvStaminaColor = CreateConVar("sm_hudbars_stamina_color", "90 190 255", "Colour of the stamina bar, \"R G B\"");
    g_cvIconBleeding = CreateConVar("sm_hudbars_icon_bleeding", "♦", "Icon shown while bleeding");
    g_cvIconInfected = CreateConVar("sm_hudbars_icon_infected", "▲", "Icon shown while infected");

    /* Netprop names. Kept as cvars so a different NMRiH build can be corrected
     * without recompiling: run sm_hudbars_scan in game to find the real ones. */
    g_cvPropStamina  = CreateConVar("sm_hudbars_prop_stamina", "m_flStamina", "Netprop holding the stamina value, empty = hide the bar");
    g_cvPropBleeding = CreateConVar("sm_hudbars_prop_bleeding", "m_bIsBleeding", "Netprop holding the bleeding flag, empty = no icon");
    g_cvPropInfected = CreateConVar("sm_hudbars_prop_infected", "m_bIsInfected", "Netprop holding the infection flag, empty = no icon");

    RegConsoleCmd("sm_hud", Cmd_ToggleHud, "Toggle the health/stamina bars for yourself");
    RegAdminCmd("sm_hudbars_scan", Cmd_Scan, ADMFLAG_GENERIC, "List the netprops this build actually exposes");

    g_hHudHealth  = CreateHudSynchronizer();
    g_hHudStamina = CreateHudSynchronizer();
    g_hHudStatus  = CreateHudSynchronizer();

    g_cvInterval.AddChangeHook(OnIntervalChanged);

    AutoExecConfig(true, "nmrih_hudbars");

    /* OnConfigsExecuted does not fire when the plugin is loaded mid-map. */
    RestartTimer();
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

void DrawFor(int client)
{
    char full[16];
    char empty[16];
    g_cvCharFull.GetString(full, sizeof(full));
    g_cvCharEmpty.GetString(empty, sizeof(empty));

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

    SetHudTextParams(x, y, hold, r, g, b, 255, 0, 0.0, 0.0, 0.0);
    if (ShowSyncHudText(client, g_hHudHealth, "HP [%s]", healthBar) < 0)
    {
        /* The game has no usable HudMsg user message: fall back to hint text,
         * which cannot be positioned or coloured but at least shows up. */
        DrawFallback(client, healthBar, cells, full, empty, inside);
        return;
    }

    float stamina = ReadProp(client, g_cvPropStamina);
    if (stamina >= 0.0)
    {
        float staminaFraction = Fraction(stamina, g_cvStaminaMax.FloatValue);

        char staminaBar[256];
        BuildBar(staminaBar, sizeof(staminaBar), staminaFraction, cells, full, empty,
            RoundToNearest(stamina), inside);

        ParseColour(g_cvStaminaColor, r, g, b);
        SetHudTextParams(x, y + line, hold, r, g, b, 255, 0, 0.0, 0.0, 0.0);
        ShowSyncHudText(client, g_hHudStamina, "SP [%s]", staminaBar);
    }

    DrawStatus(client, x, y + line * 2.0, hold);
}

void DrawStatus(int client, float x, float y, float hold)
{
    bool bleeding = ReadFlag(client, g_cvPropBleeding);
    bool infected = ReadFlag(client, g_cvPropInfected);

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

    SetHudTextParams(x, y, hold, r, g, b, 255, 0, 0.0, 0.0, 0.0);
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

    char line[320];
    Format(line, sizeof(line), "HP [%s]", healthBar);

    float stamina = ReadProp(client, g_cvPropStamina);
    if (stamina >= 0.0)
    {
        char staminaBar[256];
        BuildBar(staminaBar, sizeof(staminaBar), Fraction(stamina, g_cvStaminaMax.FloatValue),
            cells, full, empty, RoundToNearest(stamina), inside);
        Format(line, sizeof(line), "%s\nSP [%s]", line, staminaBar);
    }

    char icon[16];
    if (ReadFlag(client, g_cvPropBleeding))
    {
        g_cvIconBleeding.GetString(icon, sizeof(icon));
        Format(line, sizeof(line), "%s\n%s BLEEDING", line, icon);
    }
    if (ReadFlag(client, g_cvPropInfected))
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

    if (numberStart == -1)
    {
        Format(out, maxlen, "%s %d", out, value);
    }
}

/* Green when healthy, through yellow, to red when nearly dead. */
void HealthColour(float fraction, int &r, int &g, int &b)
{
    if (fraction <= 0.5)
    {
        float t = fraction / 0.5;
        r = 230;
        g = RoundToNearest(55.0 + t * 150.0);
        b = RoundToNearest(50.0 + t * 20.0);
    }
    else
    {
        float t = (fraction - 0.5) / 0.5;
        r = RoundToNearest(230.0 - t * 160.0);
        g = RoundToNearest(205.0 + t * 15.0);
        b = RoundToNearest(70.0 + t * 30.0);
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

/* Reads a netprop as a float, whatever its actual type is. -1.0 = not available. */
float ReadProp(int client, ConVar convar)
{
    char prop[64];
    convar.GetString(prop, sizeof(prop));
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
bool ReadFlag(int client, ConVar convar)
{
    return ReadProp(client, convar) > 0.0;
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
    if (client == 0 || !IsClientInGame(client))
    {
        ReplyToCommand(client, "[SM] Run this in game, the props are read off your own player.");
        return Plugin_Handled;
    }

    char netclass[64];
    GetEntityNetClass(client, netclass, sizeof(netclass));
    ReplyToCommand(client, "[SM] Network class: %s", netclass);

    ScanGroup(client, netclass, "stamina", g_sStaminaCandidates, sizeof(g_sStaminaCandidates));
    ScanGroup(client, netclass, "bleeding", g_sBleedingCandidates, sizeof(g_sBleedingCandidates));
    ScanGroup(client, netclass, "infection", g_sInfectedCandidates, sizeof(g_sInfectedCandidates));

    ReplyToCommand(client, "[SM] Nothing useful above? Run sm_dump_netprops_xml props.xml and search it for the class shown here.");
    return Plugin_Handled;
}

void ScanGroup(int client, const char[] netclass, const char[] label,
    const char[][] candidates, int count)
{
    ReplyToCommand(client, "[SM] --- %s ---", label);

    bool found = false;
    for (int i = 0; i < count; i++)
    {
        if (!HasEntProp(client, Prop_Send, candidates[i]))
        {
            continue;
        }

        found = true;

        PropFieldType type;
        if (FindSendPropInfo(netclass, candidates[i], type) != -1 && type == PropField_Float)
        {
            ReplyToCommand(client, "[SM]   %s = %.2f (float)", candidates[i],
                GetEntPropFloat(client, Prop_Send, candidates[i]));
        }
        else
        {
            ReplyToCommand(client, "[SM]   %s = %d (int)", candidates[i],
                GetEntProp(client, Prop_Send, candidates[i]));
        }
    }

    if (!found)
    {
        ReplyToCommand(client, "[SM]   none of the known names exist here");
    }
}
