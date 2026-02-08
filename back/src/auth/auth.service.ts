import { Injectable } from '@nestjs/common';
import { UsersService } from 'src/users/users.service';
import * as bcrypt from 'bcrypt';
import { JwtService } from '@nestjs/jwt';
import { RegisterDto } from './dto/register.dto';
import { CreateUserDto } from 'src/users/dto/create.dto';
import { SafeUserDto } from 'src/users/dto/safe.dto';
import { createHash, randomBytes } from 'crypto';
import { PrismaService } from 'src/prisma/prisma.service';

@Injectable()
export class AuthService {
    constructor(
        private readonly usersService: UsersService,
        private readonly jwtService: JwtService,
        private readonly prisma: PrismaService
    ) { }

    private generateRefreshToken(): string {
        return randomBytes(64).toString('base64url');
    }

    private hashToken(token: string): string {
        return createHash('sha256').update(token).digest('hex');
    }

    private async addRefreshToken(userId: string, tokenHash: string, expiresAt: Date): Promise<void> {
        await this.prisma.authToken.create({
            data: {
                userId,
                tokenHash,
                expiresAt,
            },
        });
    } 

    async checkUser(username: string): Promise<boolean> {
        const user = await this.usersService.findByUsername(username);
        return !!user;
    }

    async validateUser(username: string, password: string, provider: string): Promise<SafeUserDto | null> {
        const user = await this.usersService.findForLogin(username, provider);
        if (!user || !user.passwordHash) return null;

        const isPasswordValid = await bcrypt.compare(password, user.passwordHash);
        if (!isPasswordValid) return null;

        const { passwordHash, ...safeUserDto } = user;

        return safeUserDto;
    }

    async login(user: SafeUserDto) {
        const payload = { username: user.username, sub: user.id };
        const accessToken = await this.jwtService.signAsync(payload);

        const refreshToken = this.generateRefreshToken();
        const hashedRefreshToken = this.hashToken(refreshToken);

        const expiresAt = new Date(Date.now() + Number(process.env.REFRESH_TOKEN_EXPIRATION) * 24 * 60 * 60 * 1000);

        await this.addRefreshToken(user.id, hashedRefreshToken, expiresAt);

        return {
            user,
            accessToken,
            refreshToken
        }
    }

    async register(dto: RegisterDto, provider: string) {
        const createDto: CreateUserDto = {
            username: dto.username,
            passwordHash: await bcrypt.hash(dto.password, 10),
            provider,
        };

        const user = await this.usersService.create(createDto);
        return this.login(user);
    }
}
