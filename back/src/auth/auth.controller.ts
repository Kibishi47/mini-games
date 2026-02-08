import { Body, Controller, HttpCode, HttpException, HttpStatus, Post } from '@nestjs/common';
import { AuthService } from './auth.service';
import { LoginDto } from './dto/login.dto';
import { RegisterDto } from './dto/register.dto';

@Controller('auth')
export class AuthController {
    constructor(private readonly authService: AuthService) { }

    @HttpCode(HttpStatus.OK)
    @Post('login')
    async login(@Body() dto: LoginDto) {
        const user = await this.authService.validateUser(dto.username, dto.password);

        if (!user) throw new HttpException('Invalid credentials', HttpStatus.UNAUTHORIZED);

        return this.authService.login(user);
    }

    @HttpCode(HttpStatus.OK)
    @Post('register')
    async register(@Body() dto: RegisterDto) {
        const existingUser = await this.authService.checkUser(dto.username);
        if (existingUser) throw new HttpException('Username already taken', HttpStatus.BAD_REQUEST);
        
        return this.authService.register(dto);
    }
}
