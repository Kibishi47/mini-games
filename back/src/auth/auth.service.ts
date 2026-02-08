import { Injectable } from '@nestjs/common';
import { UsersService } from 'src/users/users.service';
import * as bcrypt from 'bcrypt';
import { JwtService } from '@nestjs/jwt';
import { RegisterDto } from './dto/register.dto';
import { CreateUserDto } from 'src/users/dto/create.dto';
import { User } from 'src/generated/prisma/client';
import { SafeUserDto } from 'src/users/dto/safe.dto';

@Injectable()
export class AuthService {
    constructor(
        private readonly usersService: UsersService,
        private readonly jwtService: JwtService
    ) { }

    async checkUser(username: string): Promise<boolean> {
        const user = await this.usersService.findByUsername(username);
        return !!user;
    }

    async validateUser(username: string, password: string): Promise<SafeUserDto | null> {
        const user = await this.usersService.findForLogin(username);
        if (!user || !user.passwordHash) return null;

        const isPasswordValid = await bcrypt.compare(password, user.passwordHash);
        if (!isPasswordValid) return null;

        const { passwordHash, ...safeUserDto } = user;

        return safeUserDto;
    }

    async login(user: SafeUserDto) {
        const payload = { username: user.username, sub: user.id };
        const accessToken = await this.jwtService.signAsync(payload);

        return {
            user,
            accessToken
        }
    }

    async register(dto: RegisterDto) {
        const createDto: CreateUserDto = {
            username: dto.username,
            passwordHash: await bcrypt.hash(dto.password, 10),
        };

        const user = await this.usersService.create(createDto);
        return this.login(user);
    }
}
