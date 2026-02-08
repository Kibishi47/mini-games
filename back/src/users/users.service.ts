import { Injectable } from '@nestjs/common';
import { PrismaService } from 'src/prisma/prisma.service';
import type { SafeUserDto } from './dto/safe.dto';
import { CreateUserDto } from './dto/create.dto';
import { UserWhereInput } from 'src/generated/prisma/models/User';

@Injectable()
export class UsersService {
  constructor(private readonly prisma: PrismaService) {}

  public readonly safeSelect = {
    id: true,
    username: true,
    provider: true,
    providerId: true,
    displayName: true,
    avatarUrl: true,
    createdAt: true,
  } as const;

  async exists(where: UserWhereInput): Promise<boolean> {
    const user = await this.prisma.user.findFirst({
      where,
      select: { id: true },
    });
    return !!user;
  }

  async findUserForLogin(where: UserWhereInput): Promise<(SafeUserDto & { passwordHash: string | null }) | null> {
    return this.prisma.user.findFirst({
      where,
      select: {
        ...this.safeSelect,
        passwordHash: true,
      },
    }) as any;
  }

  async createUser(dto: CreateUserDto): Promise<SafeUserDto> {
    return this.prisma.user.create({
      data: dto,
      select: this.safeSelect,
    });
  }

  async findById(id: string): Promise<SafeUserDto | null> {
    return this.prisma.user.findUnique({
      where: { id },
      select: this.safeSelect,
    });
  }
}