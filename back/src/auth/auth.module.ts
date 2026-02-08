import { Module } from '@nestjs/common';

import { AuthService } from './auth.service';
import { AuthController } from './auth.controller';

import { UsersModule } from 'src/users/users.module';
import { PrismaService } from 'src/prisma/prisma.service';


@Module({
  providers: [AuthService, PrismaService],
  imports: [
    UsersModule
  ],
  controllers: [AuthController]
})
export class AuthModule { }
