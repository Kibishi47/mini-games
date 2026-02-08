import { Injectable } from '@nestjs/common';
import { User } from 'src/generated/prisma/client';
import { PrismaService } from 'src/prisma/prisma.service';
import { CreateUserDto } from './dto/create.dto';
import { SafeUserDto } from './dto/safe.dto';

@Injectable()
export class UsersService {
    constructor(private prisma: PrismaService) {}

    async findForLogin(username: string): Promise<User | null> {
        return this.prisma.user.findUnique({
            where: {
                username,
            },
        });
    }

    async findByUsername(username: string): Promise<SafeUserDto | null> {
        return this.prisma.user.findUnique({
            select: {
                id: true,
                username: true,
                displayName: true,
                avatarUrl: true,
                createdAt: true,
            },
            where: {
                username,
            },
        });
    }

    async create(dto: CreateUserDto): Promise<SafeUserDto> {
        const user = await this.prisma.user.create({
            select: {
                id: true,
                username: true,
                displayName: true,
                avatarUrl: true,
                createdAt: true,
            },
            data: dto,
        });
        return user;
    }
}
