export type User = {
    id: string
    username: string
    provider: string
    avatarUrl: string
}

export type AuthResponse = {
  user: User;
  accessToken: string;
};

export type LoginPayload = {
  username: string;
  password: string;
};

export type RegisterPayload = {
  username: string;
  password: string;
};