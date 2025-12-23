import apiClient, { setAuthToken } from './api';

export interface LoginResponse {
    data: {
        access_token: string;
        refresh_token: string;
        user: {
            id: string;
            email: string;
            full_name: string;
        };
    };
}

export const login = async (email: string, password: string): Promise<LoginResponse> => {
    try {
        const response = await apiClient.post<LoginResponse>('/auth/login', {
            email,
            password,
        });

        if (response.data.data.access_token) {
            setAuthToken(response.data.data.access_token);
        }

        return response.data;
    } catch (error) {
        throw error;
    }
};

export const logout = async () => {
    try {
        await apiClient.post('/auth/logout', {});
    } finally {
        setAuthToken('');
    }
}
