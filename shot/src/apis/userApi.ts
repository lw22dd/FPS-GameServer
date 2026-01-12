import type { Result } from "../models/Result";
import type { User } from "../models/UserModel";
import Axios from "../utils/axios";

export default class UserApi {
    public static async login(username: string, password: string): Promise<Result<string | null>> {
        return await Axios.post('/user/login', { username, password });
    }

    public static async register(user: User): Promise<Result<string | null>> {
        return await Axios.post('/user/register', user);
    }

    public static async logout(): Promise<Result<boolean>> {
        return await Axios.post('/user/logout');
    }
    
    public static async checkUsernameExists(username: string): Promise<Result<boolean>> {
        return await Axios.get('/user/check-username', { params: { username } });
    }
    
    public static async test(str?: string ): Promise<Result<string | null>> {
        return await Axios.get('/user/test', { params: { str } });
    }
}