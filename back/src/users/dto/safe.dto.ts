export class SafeUserDto {
    id: string;
    username: string;
    provider: string;
    providerId: string | null;
    displayName: string | null;
    avatarUrl: string | null;
    createdAt: Date;
}