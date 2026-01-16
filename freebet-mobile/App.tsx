/**
 * FreeBet.guru Mobile Admin App
 *
 * Monochrome Web-Inspired Edition for React Native
 */

import React, { useCallback, useEffect, useState } from 'react';

import {
  ScrollView,
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
} from 'react-native';

import { SafeAreaView } from 'react-native-safe-area-context';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { encode as b64encode } from 'base-64';
import { StatusBar as ExpoStatusBar } from 'expo-status-bar';
import { MaterialIcons } from '@expo/vector-icons';
import * as SplashScreen from 'expo-splash-screen';

SplashScreen.preventAutoHideAsync();

export default function App() {
  // ===== СОСТОЯНИЕ ПРИЛОЖЕНИЯ =====
  const [theme, setTheme] = useState<'light' | 'dark'>('dark');
  const [serverUrl, setServerUrl] = useState('');
  const [login, setLogin] = useState('');
  const [pass, setPass] = useState('');
  const [lastResult, setLastResult] = useState<any>(null);
  const [lastRequestTime, setLastRequestTime] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'actions' | 'settings'>('actions');
  const [appIsReady, setAppIsReady] = useState(false);
  const [blinking, setBlinking] = useState(true);

  // ===== ИНИЦИАЛИЗАЦИЯ ПРИЛОЖЕНИЯ =====
  useEffect(() => {
    async function prepare() {
      try {
        await MaterialIcons.loadFont();
        const savedServerUrl = await AsyncStorage.getItem('serverUrl');
        const savedLogin = await AsyncStorage.getItem('login');
        const savedPassword = await AsyncStorage.getItem('pass');
        const savedTheme = await AsyncStorage.getItem('theme');

        if (savedServerUrl) setServerUrl(savedServerUrl);
        if (savedLogin) setLogin(savedLogin);
        if (savedPassword) setPass(savedPassword);
        if (savedTheme) setTheme(savedTheme as 'light' | 'dark');

        await performHealthCheck(
          savedServerUrl || serverUrl,
          savedLogin || login,
          savedPassword || pass
        );
        await new Promise(resolve => setTimeout(resolve, 500));
      } catch (error) {
        console.warn('Ошибка инициализации:', error);
      } finally {
        setAppIsReady(true);
      }
    }
    prepare();
  }, []);

  useEffect(() => {
    if (!lastResult?.ok) return;
    const interval = setInterval(() => {
      setBlinking(prev => !prev);
    }, 1000);
    return () => clearInterval(interval);
  }, [lastResult?.ok]);

  const onLayoutRootView = useCallback(async () => {
    if (appIsReady) {
      await SplashScreen.hideAsync();
    }
  }, [appIsReady]);

  if (!appIsReady) {
    return null;
  }

  // ===== ФУНКЦИИ API =====
  async function performHealthCheck(url: string, user: string, password: string) {
    const startTime = new Date();
    try {
      const fullUrl = url.replace(/\/$/, '') + '/api/health';
      const auth = 'Basic ' + b64encode(`${user}:${password}`);
      const response = await fetch(fullUrl, {
        method: 'GET',
        headers: {
          Authorization: auth,
          Accept: 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const json = await response.json();
      const endTime = new Date();
      const responseTime = endTime.getTime() - startTime.getTime();

      setLastResult({
        task: 'health',
        ok: json.ok,
        ms: responseTime,
        ...json
      });

    } catch (error: any) {
      setLastResult({
        task: 'health',
        ok: false,
        error: error?.message || 'Server unreachable',
        ms: 0,
      });
    } finally {
      const endTime = new Date();
      setLastRequestTime(endTime.toLocaleString('en-GB'));
    }
  }

  async function saveSettings() {
    await AsyncStorage.setItem('serverUrl', serverUrl);
    await AsyncStorage.setItem('login', login);
    await AsyncStorage.setItem('pass', pass);
    Alert.alert('Done', 'Settings saved');
  }

  const toggleTheme = async () => {
    const newTheme = theme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
    await AsyncStorage.setItem('theme', newTheme);
  };

  async function call(path: string, method = 'POST') {
    const startTime = new Date();
    try {
      const fullUrl = serverUrl.replace(/\/$/, '') + path;
      const auth = 'Basic ' + b64encode(`${login}:${pass}`);
      const response = await fetch(fullUrl, {
        method,
        headers: {
          Authorization: auth,
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
        body: method === 'POST' ? '{}' : undefined,
      });

      const json = await response.json();
      const endTime = new Date();
      const responseTime = endTime.getTime() - startTime.getTime();

      setLastResult({
        ...json,
        task: path.split('/').pop(),
        ms: responseTime,
      });

    } catch (error: any) {
      setLastResult({
        task: 'api',
        ok: false,
        error: error?.message || 'Connection error',
        ms: 0,
      });
    } finally {
      const endTime = new Date();
      setLastRequestTime(endTime.toLocaleString('en-GB'));
    }
  }

  const isDark = theme === 'dark';
  const dynamicStyles = getStyles(isDark);

  // ===== ФУНКЦИИ РЕНДЕРА =====
  const renderResult = () => {
    if (!lastResult) return <Text style={dynamicStyles.noResult}>—</Text>;

    const { task, ok, error, uptime, time, apiStats } = lastResult;
    const isSuccess = ok === true;
    const statusColor = isSuccess ? (isDark ? '#fff' : '#000') : (isDark ? '#888' : '#666');

    if (task === 'health' && isSuccess) {
      return (
        <View style={dynamicStyles.healthCard}>
          <View style={dynamicStyles.statsRow}>
            <View style={dynamicStyles.statItem}>
              <Text style={dynamicStyles.statNumber}>{lastResult.users_count || 0}</Text>
              <Text style={dynamicStyles.statLabel}>USERS</Text>
            </View>
            <View style={dynamicStyles.statDivider} />
            <View style={dynamicStyles.statItem}>
              <Text style={dynamicStyles.statNumber}>{lastResult.bets_count || 0}</Text>
              <Text style={dynamicStyles.statLabel}>BETS</Text>
            </View>
            <View style={dynamicStyles.statDivider} />
            <View style={dynamicStyles.statItem}>
              <Text style={dynamicStyles.statNumber}>{lastResult.matches_count || 0}</Text>
              <Text style={dynamicStyles.statLabel}>MATCHES</Text>
            </View>
          </View>

          <View style={dynamicStyles.infoContainer}>
            <View style={dynamicStyles.infoRow}>
              <Text style={dynamicStyles.infoLabel}>Server Time:</Text>
              <Text style={dynamicStyles.infoValue}>{time || '—'}</Text>
            </View>
            <View style={dynamicStyles.infoRow}>
              <Text style={dynamicStyles.infoLabel}>Latency:</Text>
              <Text style={dynamicStyles.infoValue}>{lastResult.ms}ms</Text>
            </View>
          </View>

          <View style={dynamicStyles.bottomContainer}>
            <Text style={dynamicStyles.uptimeText}>Uptime: {uptime}s</Text>
            <Text style={dynamicStyles.versionText}>v{lastResult.version || '1.0.0'}</Text>
          </View>
        </View>
      );
    }

    return (
      <View style={dynamicStyles.resultCard}>
        <Text style={dynamicStyles.resultTask}>{(task || 'API').toUpperCase()}</Text>
        <View style={dynamicStyles.resultRow}>
          <Text style={[dynamicStyles.resultStatus, { color: statusColor }]}>
            {isSuccess ? 'SUCCESS' : 'ERROR'}
          </Text>
          <Text style={dynamicStyles.resultTime}>{lastResult.ms}ms</Text>
        </View>
        {!isSuccess && error && (
          <Text style={dynamicStyles.errorMessage}>{error}</Text>
        )}
        {apiStats && (
          <View style={dynamicStyles.apiStatsContainer}>
            <Text style={dynamicStyles.apiStats}>Requests used: {apiStats.requests_used}</Text>
            <Text style={dynamicStyles.apiStats}>Remaining: {apiStats.requests_remaining}</Text>
          </View>
        )}
      </View>
    );
  };

  return (
    <SafeAreaView style={dynamicStyles.safeArea} onLayout={onLayoutRootView}>
      <ExpoStatusBar style={isDark ? 'light' : 'dark'} backgroundColor={isDark ? '#000' : '#fff'} />
      
      <View style={dynamicStyles.container}>
        <View style={dynamicStyles.titleContainer}>
          <MaterialIcons name="alarm" size={32} color={isDark ? '#fff' : '#000'} />
          <Text style={dynamicStyles.title}>Freebet</Text>
        </View>

        <View style={dynamicStyles.tabContainer}>
          <TouchableOpacity
            style={[dynamicStyles.tab, activeTab === 'actions' && dynamicStyles.activeTab]}
            onPress={() => setActiveTab('actions')}
          >
            <Text style={[dynamicStyles.tabText, activeTab === 'actions' && dynamicStyles.activeTabText]}>
              Actions
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[dynamicStyles.tab, activeTab === 'settings' && dynamicStyles.activeTab]}
            onPress={() => setActiveTab('settings')}
          >
            <Text style={[dynamicStyles.tabText, activeTab === 'settings' && dynamicStyles.activeTabText]}>
              Settings
            </Text>
          </TouchableOpacity>
        </View>

        <ScrollView contentContainerStyle={dynamicStyles.scrollContent} showsVerticalScrollIndicator={false}>
          {activeTab === 'actions' && (
            <>
              <View style={dynamicStyles.block}>
                <TouchableOpacity style={dynamicStyles.button} onPress={() => call('/api/odds/sync')}>
                  <Text style={dynamicStyles.buttonText}>Get & Sync Odds</Text>
                </TouchableOpacity>

                <TouchableOpacity style={dynamicStyles.button} onPress={() => call('/api/scores/sync')}>
                  <Text style={dynamicStyles.buttonText}>Get & Sync Scores</Text>
                </TouchableOpacity>

                <TouchableOpacity style={dynamicStyles.button} onPress={() => call('/api/calc')}>
                  <Text style={dynamicStyles.buttonText}>Calculate All Bets</Text>
                </TouchableOpacity>

                <TouchableOpacity style={[dynamicStyles.button, dynamicStyles.buttonSecondary]} onPress={() => performHealthCheck(serverUrl, login, pass)}>
                  <Text style={dynamicStyles.buttonTextSecondary}>API Health Check</Text>
                </TouchableOpacity>
              </View>

              <View style={dynamicStyles.block}>
                <View style={dynamicStyles.statusSection}>
                   <View style={[dynamicStyles.statusIndicator, { opacity: blinking ? 1 : 0.5 }]}>
                      <View style={dynamicStyles.statusDot} />
                   </View>
                   <Text style={dynamicStyles.statusLabelText}>
                      SERVER: {lastResult?.ok ? 'ONLINE' : 'OFFLINE'}
                   </Text>
                </View>
                {renderResult()}
              </View>
            </>
          )}

          {activeTab === 'settings' && (
            <View style={dynamicStyles.block}>
              <Text style={dynamicStyles.label}>Server URL</Text>
              <TextInput
                style={dynamicStyles.input}
                value={serverUrl}
                onChangeText={setServerUrl}
                placeholder=""
                placeholderTextColor={isDark ? '#666' : '#999'}
                autoCapitalize="none"
              />

              <Text style={dynamicStyles.label}>Login</Text>
              <TextInput
                style={dynamicStyles.input}
                value={login}
                onChangeText={setLogin}
                placeholder=""
                placeholderTextColor={isDark ? '#666' : '#999'}
                autoCapitalize="none"
              />

              <Text style={dynamicStyles.label}>Password</Text>
              <TextInput
                style={dynamicStyles.input}
                value={pass}
                onChangeText={setPass}
                secureTextEntry
                placeholder=""
                placeholderTextColor={isDark ? '#666' : '#999'}
              />

              <TouchableOpacity style={dynamicStyles.button} onPress={saveSettings}>
                <Text style={dynamicStyles.buttonText}>Save Settings</Text>
              </TouchableOpacity>
            </View>
          )}
        </ScrollView>

        <View style={dynamicStyles.footer}>
          <TouchableOpacity style={dynamicStyles.themeToggle} onPress={toggleTheme}>
             <MaterialIcons name={isDark ? 'light-mode' : 'dark-mode'} size={20} color={isDark ? '#fff' : '#000'} />
             <Text style={dynamicStyles.themeToggleText}>SWITCH TO {isDark ? 'LIGHT' : 'DARK'} MODE</Text>
          </TouchableOpacity>
          <Text style={dynamicStyles.footerText}>v1.0.0 // MONOCHROME</Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const getStyles = (isDark: boolean) => StyleSheet.create({
  safeArea: { flex: 1, backgroundColor: isDark ? '#000' : '#fff' },
  container: { flex: 1, backgroundColor: isDark ? '#000' : '#fff' },
  titleContainer: {
    flexDirection: 'row',
    paddingVertical: 40,
    justifyContent: 'center',
    alignItems: 'center',
    borderBottomWidth: 1,
    borderBottomColor: isDark ? '#222' : '#eee',
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: isDark ? '#fff' : '#000',
    marginLeft: 12,
    letterSpacing: 2,
    textTransform: 'uppercase',
  },
  tabContainer: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: isDark ? '#222' : '#eee',
  },
  tab: {
    flex: 1,
    paddingVertical: 16,
    alignItems: 'center',
  },
  activeTab: {
    borderBottomWidth: 2,
    borderBottomColor: isDark ? '#fff' : '#000',
  },
  tabText: {
    color: isDark ? '#666' : '#999',
    fontSize: 14,
    fontWeight: '600',
    textTransform: 'uppercase',
  },
  activeTabText: { color: isDark ? '#fff' : '#000' },
  scrollContent: { padding: 20 },
  block: { marginBottom: 24 },
  button: {
    backgroundColor: isDark ? '#fff' : '#000',
    paddingVertical: 16,
    borderRadius: 4,
    alignItems: 'center',
    marginTop: 12,
  },
  buttonSecondary: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: isDark ? '#333' : '#ddd',
  },
  buttonText: {
    color: isDark ? '#000' : '#fff',
    fontSize: 14,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  buttonTextSecondary: {
    color: isDark ? '#fff' : '#000',
    fontSize: 14,
    fontWeight: '600',
    textTransform: 'uppercase',
  },
  label: {
    color: isDark ? '#888' : '#666',
    fontSize: 12,
    fontWeight: '600',
    marginTop: 16,
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  input: {
    backgroundColor: isDark ? '#111' : '#f9f9f9',
    color: isDark ? '#fff' : '#000',
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 14,
    borderWidth: 1,
    borderColor: isDark ? '#222' : '#eee',
    borderRadius: 4,
    marginTop: 8,
  },
  resultCard: {
    backgroundColor: isDark ? '#111' : '#f9f9f9',
    padding: 20,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: isDark ? '#222' : '#eee',
  },
  resultTask: { color: isDark ? '#666' : '#999', fontSize: 10, fontWeight: '700', letterSpacing: 1 },
  resultRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 },
  resultStatus: { fontSize: 18, fontWeight: '700' },
  resultTime: { color: isDark ? '#666' : '#999', fontSize: 12 },
  errorMessage: { color: isDark ? '#888' : '#666', marginTop: 8, fontSize: 12 },
  apiStatsContainer: { marginTop: 12, borderTopWidth: 1, borderTopColor: isDark ? '#222' : '#eee', paddingTop: 12 },
  apiStats: { color: isDark ? '#888' : '#666', fontSize: 12, marginBottom: 4 },
  healthCard: {
    backgroundColor: isDark ? '#111' : '#f9f9f9',
    padding: 20,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: isDark ? '#222' : '#eee',
  },
  statusSection: { flexDirection: 'row', alignItems: 'center', marginBottom: 12, paddingLeft: 4 },
  statusIndicator: { marginRight: 10 },
  statusDot: { width: 10, height: 10, borderRadius: 5, backgroundColor: isDark ? '#fff' : '#000' },
  statusLabelText: { color: isDark ? '#fff' : '#000', fontSize: 11, fontWeight: '700', letterSpacing: 1 },
  statsRow: { flexDirection: 'row', flex: 1, justifyContent: 'space-around', marginBottom: 15 },
  statItem: { alignItems: 'center' },
  statNumber: { color: isDark ? '#fff' : '#000', fontSize: 20, fontWeight: '700' },
  statLabel: { color: isDark ? '#666' : '#999', fontSize: 9, fontWeight: '700', marginTop: 4 },
  statDivider: { width: 1, height: 20, backgroundColor: isDark ? '#222' : '#eee' },
  infoContainer: { borderTopWidth: 1, borderTopColor: isDark ? '#222' : '#eee', paddingTop: 15 },
  infoRow: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 8 },
  infoLabel: { color: isDark ? '#666' : '#999', fontSize: 12 },
  infoValue: { color: isDark ? '#fff' : '#000', fontSize: 12, fontWeight: '600' },
  bottomContainer: { flexDirection: 'row', justifyContent: 'space-between', marginTop: 15 },
  uptimeText: { color: isDark ? '#444' : '#ccc', fontSize: 10 },
  versionText: { color: isDark ? '#444' : '#ccc', fontSize: 10 },
  noResult: { color: isDark ? '#333' : '#ddd', textAlign: 'center', padding: 20 },
  footer: { padding: 20, alignItems: 'center', borderTopWidth: 1, borderTopColor: isDark ? '#222' : '#eee' },
  themeToggle: { flexDirection: 'row', alignItems: 'center', marginBottom: 15, padding: 10 },
  themeToggleText: { color: isDark ? '#fff' : '#000', fontSize: 10, fontWeight: '700', marginLeft: 8, letterSpacing: 1 },
  footerText: { color: isDark ? '#222' : '#ddd', fontSize: 10, letterSpacing: 2 },
});
