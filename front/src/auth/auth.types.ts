export type User = {
  id: string
  username: string
  displayName: string
  avatarUrl: string
  createdAt: string
}

export type AuthResponse = {
  accessToken: string;
  expiresAt: string;
};

export type MeResponse = {
  user: User;
};

export type LoginPayload = {
  username: string;
  password: string;
};

export type RegisterPayload = {
  username: string;
  password: string;
};