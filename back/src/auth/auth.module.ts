import { Module } from '@nestjs/common';
import { AuthService } from './auth.service';

import { UsersModule } from 'src/users/users.module';

import { JwtModule } from '@nestjs/jwt';
import { AuthController } from './auth.controller';

@Module({
  providers: [AuthService],
  imports: [
    UsersModule
  ],
  controllers: [AuthController]
})
export class AuthModule { }
