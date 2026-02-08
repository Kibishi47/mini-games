import { Injectable } from '@nestjs/common';
import { User } from 'src/generated/prisma/client';
import { PrismaService } from 'src/prisma/prisma.service';
import { CreateUserDto } from './dto/create.dto';
import { SafeUserDto } from './dto/safe.dto';

const safeColumns = {
    id: true,
    username: true,
    provider: true,
    providerId: true,
    displayName: true,
    avatarUrl: true,
    createdAt: true,
}

@Injectable()
export class UsersService {
    constructor(private readonly prisma: PrismaService) {}

    async findForLogin(username: string, provider: string): Promise<User | null> {
        return this.prisma.user.findFirst({
            where: {
                username,
                provider,
            },
        });
    }

    async findByUsername(username: string): Promise<SafeUserDto | null> {
        return this.prisma.user.findFirst({
            select: safeColumns,
            where: {
                username,
            },
        });
    }

    async create(dto: CreateUserDto): Promise<SafeUserDto> {
        const user = await this.prisma.user.create({
            select: safeColumns,
            data: dto,
        });
        return user;
    }
}
