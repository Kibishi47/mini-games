export class SafeUserDto {
    id: string;
    username: string;
    displayName: string | null;
    avatarUrl: string | null;
    createdAt: Date;
}