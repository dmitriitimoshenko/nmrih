/**
 * Plugin: nmrih_hud_stats.sp
 * Назначение: Отслеживание убийств зомби и вывод на HUD таблицы со статистикой.
 *
 * Функционал:
 * - Отслеживание убийств зомби с разделением на убийства холодным оружием (название оружия начинается с "me_")
 *   и убийства огнестрельным оружием (остальные).
 * - Вывод слева на экране HUD-таблицы, где для каждого игрока отображается:
 *     Никнейм | Меле | Огнестрел | Всего
 * - Сброс статистики происходит в начале нового раунда (обрабатывается событие "nmrih_round_begin").
 */

#include <sourcemod>
#include <sdktools>
#include <halflife>

int g_MeleeKills[MAXPLAYERS+1];
int g_GunKills[MAXPLAYERS+1];

public void OnPluginStart()
{
    // Перехват события убийства NPC (зомби)
    HookEvent("npc_killed", Event_NPCKilled, EventHookMode_Post);
    // Перехват события начала раунда для сброса статистики
    HookEvent("nmrih_round_begin", Event_RoundBegin, EventHookMode_Post);
    
    PrintToServer("NMRIH HUD Stats Plugin успешно загружен!");
}

/**
 * Обработчик события npc_killed.
 * Если убийца является игроком, определяется тип оружия и увеличивается соответствующий счетчик.
 */
public Action Event_NPCKilled(Event event, const char[] name, bool dontBroadcast)
{
    int attacker = event.GetInt("attacker");
    if (attacker <= 0 || !IsClientInGame(attacker))
    {
        return Plugin_Continue;
    }
    
    char weapon[64];
    event.GetString("weapon", weapon, sizeof(weapon));
    
    // Если имя оружия начинается с "me_", считаем, что это холодное оружие
    if (StrContains(weapon, "me_", false) == 0)
    {
        g_MeleeKills[attacker]++;
    }
    else
    {
        g_GunKills[attacker]++;
    }
    
    UpdateHUDForAll();
    return Plugin_Continue;
}

/**
 * Обработчик события начала раунда.
 * Сбрасывает счетчики убийств для всех игроков.
 */
public Action Event_RoundBegin(Event event, const char[] name, bool dontBroadcast)
{
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsClientInGame(i))
        {
            g_MeleeKills[i] = 0;
            g_GunKills[i] = 0;
        }
    }
    
    UpdateHUDForAll();
    return Plugin_Continue;
}

/**
 * Функция обновления HUD-таблицы для всех игроков.
 */
void UpdateHUDForAll()
{
    SetHudTextParams(0.05, 0.2, 10.0, 200, 200, 200, 5, 0, 6.0, 0.5, 0.5);
    Handle sync = CreateHudSynchronizer();


    //HUDTextParams hudParams;
    //hudParams.x = 0.05;          // Расположение по горизонтали (слева)
    //hudParams.y = 0.2;           // Расположение по вертикали
    //hudParams.holdTime = 10.0;   // Время отображения текста
    //hudParams.fadeInTime = 0.5;
    //hudParams.fadeOutTime = 0.5;
    //hudParams.channel = 4;       // Канал, который будет перезаписываться новым текстом
    //hudParams.effect = 0;        // Без спецэффектов

    char hudText[2048];
    Format(hudText, sizeof(hudText), "Статистика убийств:\n");
    
    char temp[128];
    char clientName[64];
    
    // Формируем строку для каждого игрока
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsClientInGame(i))
        {
            GetClientName(i, clientName, sizeof(clientName));
            int melee = g_MeleeKills[i];
            int gun   = g_GunKills[i];
            int total = melee + gun;
            
            Format(temp, sizeof(temp), "%-16s Меле: %2d  Огнестрел: %2d  Всего: %2d\n",
                   clientName, melee, gun, total);
            
            strcopy(hudText + strlen(hudText), sizeof(hudText) - strlen(hudText), temp);
        }
    }
    
    // Отправляем обновленный HUD всем игрокам
    for (int i = 1; i <= MaxClients; i++)
    {
        if (IsClientInGame(i))
        {
            ShowSyncHudText(i, sync, hudText);
        }
    }
}
