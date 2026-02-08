import { Body, Controller, HttpCode, HttpException, HttpStatus, Post, Req, Res } from '@nestjs/common';
import type { Response } from 'express';
import { AuthService } from './auth.service';
import { LoginDto } from './dto/login.dto';
import { RegisterDto } from './dto/register.dto';

@Controller('auth')
export class AuthController {
    constructor(private readonly authService: AuthService) { }

    private setRefreshTokenCookie(response: Response, token: string) {
        response.cookie('refresh_token', token, {
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict',
            maxAge: Number(process.env.REFRESH_TOKEN_EXPIRATION) * 24 * 60 * 60 * 1000,
        });
    }

    @HttpCode(HttpStatus.OK)
    @Post('login')
    async login(@Body() dto: LoginDto, @Res({ passthrough: true }) response: Response) {
        const user = await this.authService.validateUser(dto.username, dto.password, "local");

        if (!user) throw new HttpException('Invalid credentials', HttpStatus.UNAUTHORIZED);

        const { refreshToken, ...data } = await this.authService.login(user);

        this.setRefreshTokenCookie(response, refreshToken);

        return data;
    }

    @HttpCode(HttpStatus.OK)
    @Post('register')
    async register(@Body() dto: RegisterDto, @Res({ passthrough: true }) response: Response) {
        const existingUser = await this.authService.checkUser(dto.username);
        if (existingUser) throw new HttpException('Username already taken', HttpStatus.BAD_REQUEST);

        const { refreshToken, ...data } = await this.authService.register(dto, "local");

        this.setRefreshTokenCookie(response, refreshToken);
        
        return data;
    }
}
