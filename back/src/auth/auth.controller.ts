import {
    Body,
    Controller,
    HttpCode,
    HttpStatus,
    Post,
    Req,
    Res,
} from '@nestjs/common';
import type { Request, Response } from 'express';
import { AuthService } from './auth.service';
import { LoginDto } from './dto/login.dto';
import { RegisterDto } from './dto/register.dto';

@Controller('auth')
export class AuthController {
    constructor(private readonly authService: AuthService) { }

    private setRefreshCookie(res: Response, token: string) {
        const days = Number(process.env.REFRESH_TOKEN_EXPIRATION ?? 30);
        res.cookie('refresh_token', token, {
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'lax',
            path: '/auth',
            maxAge: days * 24 * 60 * 60 * 1000,
        });
    }

    private clearRefreshCookie(res: Response) {
        res.clearCookie('refresh_token', {
            path: '/auth',
        });
    }

    private getRefreshToken(req: Request): string | undefined {
        return req.cookies?.refresh_token as string | undefined;
    }

    @HttpCode(HttpStatus.OK)
    @Post('login')
    async login(
        @Body() dto: LoginDto,
        @Req() req: Request,
        @Res({ passthrough: true }) res: Response
    ) {
        const oldRefreshToken = this.getRefreshToken(req);
        const { refreshToken, ...data } = await this.authService.loginLocal(dto, oldRefreshToken);
        this.setRefreshCookie(res, refreshToken);
        return data;
    }

    @HttpCode(HttpStatus.OK)
    @Post('register')
    async register(
        @Body() dto: RegisterDto,
        @Res({ passthrough: true }) res: Response,
    ) {
        const { refreshToken, ...data } = await this.authService.registerLocal(dto);
        this.setRefreshCookie(res, refreshToken);
        return data;
    }

    @HttpCode(HttpStatus.OK)
    @Post('refresh')
    async refresh(
        @Req() req: Request,
        @Res({ passthrough: true }) res: Response
    ) {
        const oldRefreshToken = this.getRefreshToken(req);

        const { refreshToken, ...data } =
            await this.authService.refreshSession(oldRefreshToken);

        this.setRefreshCookie(res, refreshToken);
        return data;
    }

    @HttpCode(HttpStatus.OK)
    @Post('logout')
    async logout(
        @Req() req: Request,
        @Res({ passthrough: true }) res: Response
    ) {
        const refreshToken = this.getRefreshToken(req);
        await this.authService.logout(refreshToken);
        this.clearRefreshCookie(res);
        return { ok: true };
    }
}