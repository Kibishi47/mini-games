export class CreateUserDto {
    username: string;
    passwordHash: string;
    provider: string = "local";
    providerId?: string | null = null;
    displayName?: string | null = null;
    avatarUrl?: string | null = null;
}