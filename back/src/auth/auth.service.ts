import {
    BadRequestException,
    Injectable,
    UnauthorizedException,
} from '@nestjs/common';
import * as bcrypt from 'bcrypt';
import { JwtService } from '@nestjs/jwt';
import { PrismaService } from 'src/prisma/prisma.service';
import { UsersService } from 'src/users/users.service';
import { createHash, randomBytes } from 'crypto';
import { LoginDto } from './dto/login.dto';
import { RegisterDto } from './dto/register.dto';
import type { SafeUserDto } from 'src/users/dto/safe.dto';
import { CreateUserDto } from 'src/users/dto/create.dto';

type AuthResult = {
    user: SafeUserDto;
    accessToken: string;
    refreshToken: string;
};

@Injectable()
export class AuthService {
    constructor(
        private readonly usersService: UsersService,
        private readonly jwtService: JwtService,
        private readonly prisma: PrismaService,
    ) { }

    // ===== Public API =====

    async loginLocal(dto: LoginDto, oldRefreshToken?: string): Promise<AuthResult> {
        const user = await this.validateLocalCredentials(dto.username, dto.password);

        const refreshToken = await this.rotateRefreshToken(user.id, oldRefreshToken);
        return this.issueSession(user, refreshToken);
    }

    async registerLocal(dto: RegisterDto): Promise<AuthResult> {
        const exists = await this.usersService.exists({ username: dto.username, provider: 'local' });
        if (exists) throw new BadRequestException('Username already taken');

        const passwordHash = await bcrypt.hash(dto.password, 10);

        const createDto: CreateUserDto = {
            username: dto.username,
            passwordHash,
            provider: 'local',
        };

        const user = await this.usersService.createUser(createDto);

        const refreshToken = await this.createRefreshToken(user.id);
        return this.issueSession(user, refreshToken);
    }

    async refreshSession(oldRefreshToken?: string): Promise<AuthResult> {
        if (!oldRefreshToken) throw new UnauthorizedException('No refresh token');

        const { user } = await this.validateRefreshToken(oldRefreshToken);

        // rotation: supprime l'ancien et en crée un nouveau
        const newRefreshToken = await this.rotateRefreshToken(user.id, oldRefreshToken);

        return this.issueSession(user, newRefreshToken);
    }

    async logout(refreshToken?: string): Promise<void> {
        if (!refreshToken) return;
        const hash = this.hashToken(refreshToken);
        await this.prisma.authToken.deleteMany({ where: { tokenHash: hash } });
    }

    // ===== Credentials =====

    private async validateLocalCredentials(username: string, password: string): Promise<SafeUserDto> {
        const user = await this.usersService.findUserForLogin({ username, provider: 'local' });
        if (!user?.passwordHash) throw new UnauthorizedException('Invalid credentials');

        const ok = await bcrypt.compare(password, user.passwordHash);
        if (!ok) throw new UnauthorizedException('Invalid credentials');

        const { passwordHash, ...safeUser } = user;
        return safeUser;
    }

    // ===== Session issuance =====

    private async issueSession(user: SafeUserDto, refreshToken: string): Promise<AuthResult> {
        const accessToken = await this.issueAccessToken(user);
        return { user, accessToken, refreshToken };
    }

    private async issueAccessToken(user: SafeUserDto): Promise<string> {
        const payload = { sub: user.id, username: user.username, provider: user.provider };
        return await this.jwtService.signAsync(payload);
    }

    // ===== Refresh token handling =====

    private async createRefreshToken(userId: string): Promise<string> {
        const refreshToken = this.generateRefreshToken();
        const tokenHash = this.hashToken(refreshToken);

        const days = Number(process.env.REFRESH_TOKEN_EXPIRATION ?? 30);
        const expiresAt = new Date(Date.now() + days * 24 * 60 * 60 * 1000);

        await this.prisma.authToken.create({
            data: {
                userId,
                tokenHash,
                expiresAt,
            },
        });

        return refreshToken;
    }

    private async rotateRefreshToken(userId: string, oldRefreshToken?: string): Promise<string> {

        if (!oldRefreshToken) {
            return this.createRefreshToken(userId);
        }

        const oldHash = this.hashToken(oldRefreshToken);

        // sécurité: supprimer uniquement si le token appartient bien à l'user
        const deleted = await this.prisma.authToken.deleteMany({
            where: {
                tokenHash: oldHash,
                userId,
            },
        });

        if (deleted.count === 0) {
            throw new UnauthorizedException('Invalid refresh token');
        }

        return this.createRefreshToken(userId);
    }

    private async validateRefreshToken(refreshToken: string): Promise<{ user: SafeUserDto; tokenHash: string }> {
        const tokenHash = this.hashToken(refreshToken);

        const row = await this.prisma.authToken.findUnique({
            where: { tokenHash },
            select: {
                tokenHash: true,
                expiresAt: true,
                user: { select: this.usersService.safeSelect },
            },
        });

        if (!row || row.expiresAt < new Date()) {
            throw new UnauthorizedException('Invalid refresh token');
        }

        return { user: row.user, tokenHash: row.tokenHash };
    }

    // ===== Crypto helpers =====

    private generateRefreshToken(): string {
        return randomBytes(64).toString('base64url');
    }

    private hashToken(token: string): string {
        return createHash('sha256').update(token).digest('hex');
    }
}