import React, { useState, useRef } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  FlatList,
} from 'react-native';

export default function App() {
  const [logs, setLogs] = useState([]);
  const flatListRef = useRef(null);

  const addLog = (log) => {
    const now = new Date();
    const timestamp = now.toISOString().replace(/[TZ]/g, ' ');
    setLogs((prevLogs) => [...prevLogs, `${timestamp} UTC: ${log}`]);
  };

  const sendRequest = async () => {
    const abortController = new AbortController();
    setTimeout(() => {
      abortController.abort();
    }, 2000);

    let response = undefined;
    try {
      response = await fetch(
          'https://backend-ri6qxvjyda-uw.a.run.app/api/open',
          {
            method: 'POST',
            signal: abortController.signal,
            body: JSON.stringify({
              action: 'open',
              toServer: {
                shortResponse: true,
              },
              passthrough: {
                who: 'phone',
              },
            }),
          }
      );

      if (response.ok) {
        addLog('😄 opened');
      } else {
        addLog(`😧 failed to open, response: ${JSON.stringify(response)}`);
      }
    } catch (error) {
      addLog(
          `🫤 failed to open, error: ${error.message}, response: ${JSON.stringify(
              response
          )}`
      );
    }
  };

  return (
      <SafeAreaView style={styles.container}>
        <View style={styles.logContainer}>
          <FlatList
              ref={flatListRef}
              data={logs}
              renderItem={({ item }) => (
                  <Text selectable={true} style={styles.logText}>
                    {item}
                  </Text>
              )}
              keyExtractor={(item, index) => index.toString()}
              contentContainerStyle={styles.logContentContainer}
              onContentSizeChange={() =>
                  flatListRef.current.scrollToEnd({ animated: true })
              }
          />
        </View>

        <TouchableOpacity style={styles.button} onPress={sendRequest}>
          <Text style={styles.buttonText}> Open Gate </Text>
        </TouchableOpacity>
      </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000000',
    alignItems: 'center',
    justifyContent: 'flex',
  },
  logContainer: {
    flex: 1,
    alignSelf: 'stretch',
  },
  logContentContainer: {
    padding: 10,
    flexGrow: 1, // 让内容占满可滚动区域
  },
  logText: {
    color: '#FFFFFF',
    marginBottom: 10,
    fontFamily: 'Helvetica',
    fontSize: 14,
    fontWeight: 'bold',
  },
  button: {
    alignSelf: 'stretch',
    backgroundColor: '#000000',
    borderWidth: 1,
    borderColor: '#FFFFFF',
    borderRadius: 8,
    paddingHorizontal: 20,
    paddingVertical: 10,
    marginHorizontal: 20,
    marginBottom: 20,
  },
  buttonText: {
    color: 'pink',
    fontFamily: 'Arial',
    fontSize: 16,
    fontWeight: 'bold',
    textAlign: 'center',
  },
});
