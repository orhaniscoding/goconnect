import axios from 'axios';

// Default to localhost for Android emulator (10.0.2.2 usually mapping to host 127.0.0.1)
// Adjust as needed for physical device testing
export const API_URL = 'http://10.0.2.2:8080/v1';

const apiClient = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

let authToken = '';

export const setAuthToken = (token: string) => {
    authToken = token;
    if (token) {
        apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
        delete apiClient.defaults.headers.common['Authorization'];
    }
};

export default apiClient;
