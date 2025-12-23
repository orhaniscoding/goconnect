import React, { useEffect, useState } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, Alert } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import apiClient from '../services/api';
import { logout } from '../services/auth';

const HomeScreen = ({ navigation }: any) => {
    const [networks, setNetworks] = useState<any[]>([]);

    useEffect(() => {
        fetchNetworks();
    }, []);

    const fetchNetworks = async () => {
        try {
            const response = await apiClient.get('/networks');
            if (response.data.data) {
                setNetworks(response.data.data);
            }
        } catch (error) {
            console.error(error);
            Alert.alert('Error', 'Failed to fetch networks');
        }
    };

    const handleLogout = async () => {
        await logout();
        navigation.replace('Login');
    }

    const renderItem = ({ item }: { item: any }) => (
        <TouchableOpacity style={styles.card}>
            <Text style={styles.networkName}>{item.name}</Text>
            <Text style={styles.networkCIDR}>{item.cidr}</Text>
            <View style={styles.statusBadge}>
                <Text style={styles.statusText}>Connected</Text>
                {/* Mock status for now */}
            </View>
        </TouchableOpacity>
    );

    return (
        <SafeAreaView style={styles.container}>
            <View style={styles.header}>
                <Text style={styles.title}>My Networks</Text>
                <TouchableOpacity onPress={handleLogout}>
                    <Text style={styles.logoutText}>Logout</Text>
                </TouchableOpacity>
            </View>

            <FlatList
                data={networks}
                renderItem={renderItem}
                keyExtractor={(item) => item.id}
                contentContainerStyle={styles.list}
                refreshing={false}
                onRefresh={fetchNetworks}
                ListEmptyComponent={<Text style={styles.emptyText}>No networks found</Text>}
            />
        </SafeAreaView>
    );
};

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#f2f2f7', // iOS Light Gray
    },
    header: {
        padding: 20,
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        backgroundColor: '#fff',
        borderBottomWidth: 1,
        borderBottomColor: '#e5e5ea',
    },
    title: {
        fontSize: 28,
        fontWeight: 'bold',
    },
    logoutText: {
        color: '#007AFF',
        fontSize: 16,
    },
    list: {
        padding: 16,
    },
    card: {
        backgroundColor: '#fff',
        padding: 20,
        borderRadius: 16,
        marginBottom: 16,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 8,
        elevation: 2,
    },
    networkName: {
        fontSize: 18,
        fontWeight: '600',
        marginBottom: 4,
    },
    networkCIDR: {
        fontSize: 14,
        color: '#8e8e93',
        marginBottom: 12,
    },
    statusBadge: {
        backgroundColor: '#34c759', // Green
        paddingHorizontal: 8,
        paddingVertical: 4,
        borderRadius: 6,
        alignSelf: 'flex-start',
    },
    statusText: {
        color: '#fff',
        fontSize: 12,
        fontWeight: 'bold',
    },
    emptyText: {
        textAlign: 'center',
        marginTop: 50,
        color: '#8e8e93',
    }
});

export default HomeScreen;
